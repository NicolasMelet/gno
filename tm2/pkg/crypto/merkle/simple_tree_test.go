package merkle

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/gnolang/gno/tm2/pkg/testutils"
)

type testItem []byte

func (tI testItem) Hash() []byte {
	return []byte(tI)
}

func TestSimpleProof(t *testing.T) {
	t.Parallel()

	total := 100

	items := make([][]byte, total)
	for i := range total {
		items[i] = testItem(random.RandBytes(tmhash.Size))
	}

	rootHash := SimpleHashFromByteSlices(items)

	rootHash2, proofs := SimpleProofsFromByteSlices(items)

	require.Equal(t, rootHash, rootHash2, "Unmatched root hashes: %X vs %X", rootHash, rootHash2)

	// For each item, check the trail.
	for i, item := range items {
		proof := proofs[i]

		// Check total/index
		require.Equal(t, proof.Index, i, "Unmatched indices: %d vs %d", proof.Index, i)

		require.Equal(t, proof.Total, total, "Unmatched totals: %d vs %d", proof.Total, total)

		// Verify success
		err := proof.Verify(rootHash, item)
		require.NoError(t, err, "Verification failed: %v.", err)

		// Trail too long should make it fail
		origAunts := proof.Aunts
		proof.Aunts = append(proof.Aunts, random.RandBytes(32))
		err = proof.Verify(rootHash, item)
		require.Error(t, err, "Expected verification to fail for wrong trail length")

		proof.Aunts = origAunts

		// Trail too short should make it fail
		proof.Aunts = proof.Aunts[0 : len(proof.Aunts)-1]
		err = proof.Verify(rootHash, item)
		require.Error(t, err, "Expected verification to fail for wrong trail length")

		proof.Aunts = origAunts

		// Mutating the itemHash should make it fail.
		err = proof.Verify(rootHash, testutils.MutateByteSlice(item))
		require.Error(t, err, "Expected verification to fail for mutated leaf hash")

		// Mutating the rootHash should make it fail.
		err = proof.Verify(testutils.MutateByteSlice(rootHash), item)
		require.Error(t, err, "Expected verification to fail for mutated root hash")
	}
}

func TestSimpleHashAlternatives(t *testing.T) {
	t.Parallel()

	total := 100

	items := make([][]byte, total)
	for i := range total {
		items[i] = testItem(random.RandBytes(tmhash.Size))
	}

	rootHash1 := SimpleHashFromByteSlicesIterative(items)
	rootHash2 := SimpleHashFromByteSlices(items)
	require.Equal(t, rootHash1, rootHash2, "Unmatched root hashes: %X vs %X", rootHash1, rootHash2)
}

func BenchmarkSimpleHashAlternatives(b *testing.B) {
	total := 100

	items := make([][]byte, total)
	for i := range total {
		items[i] = testItem(random.RandBytes(tmhash.Size))
	}

	b.ResetTimer()
	b.Run("recursive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = SimpleHashFromByteSlices(items)
		}
	})

	b.Run("iterative", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = SimpleHashFromByteSlicesIterative(items)
		}
	})
}

func Test_getSplitPoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		length int
		want   int
	}{
		{1, 0},
		{2, 1},
		{3, 2},
		{4, 2},
		{5, 4},
		{10, 8},
		{20, 16},
		{100, 64},
		{255, 128},
		{256, 128},
		{257, 256},
	}
	for _, tt := range tests {
		got := getSplitPoint(tt.length)
		require.Equal(t, tt.want, got, "getSplitPoint(%d) = %v, want %v", tt.length, got, tt.want)
	}
}
