package ffi

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportSector(t *testing.T) {
	challengeCount := uint64(2)
	poRepProofPartitions := uint8(10)
	proverID := [32]byte{6, 7, 8}
	randomness := [32]byte{9, 9, 9}
	sectorSize := uint64(1024)
	sectorID := uint64(42)

	ticket := SealTicket{
		BlockHeight: 0,
		TicketBytes: [32]byte{5, 4, 2},
	}

	seed := SealSeed{
		BlockHeight: 50,
		TicketBytes: [32]byte{7, 4, 2},
	}

	// initialize a sector builder
	metadataDir := requireTempDirPath(t, "metadata")
	defer os.RemoveAll(metadataDir)

	sealedSectorsDir := requireTempDirPath(t, "sealed-sectors")
	defer os.RemoveAll(sealedSectorsDir)

	stagedSectorsDir := requireTempDirPath(t, "staged-sectors")
	defer os.RemoveAll(stagedSectorsDir)

	sectorCacheRootDir := requireTempDirPath(t, "sector-cache-root-dir")
	defer os.RemoveAll(sectorCacheRootDir)

	sectorCacheDirPath := requireTempDirPath(t, "sector-cache-dir")
	defer os.RemoveAll(sectorCacheDirPath)

	stagedSectorFile := requireTempFile(t, bytes.NewReader([]byte{}), 0)
	defer stagedSectorFile.Close()

	sealedSectorFile := requireTempFile(t, bytes.NewReader([]byte{}), 0)
	defer sealedSectorFile.Close()

	unsealOutputFileA := requireTempFile(t, bytes.NewReader([]byte{}), 0)
	defer unsealOutputFileA.Close()

	unsealOutputFileB := requireTempFile(t, bytes.NewReader([]byte{}), 0)
	defer unsealOutputFileB.Close()

	unsealOutputFileC := requireTempFile(t, bytes.NewReader([]byte{}), 0)
	defer unsealOutputFileC.Close()

	unsealOutputFileD := requireTempFile(t, bytes.NewReader([]byte{}), 0)
	defer unsealOutputFileD.Close()

	// some rando bytes
	someBytes := make([]byte, 1016)
	_, err := io.ReadFull(rand.Reader, someBytes)
	require.NoError(t, err)

	// write first piece
	require.NoError(t, err)
	pieceFileA := requireTempFile(t, bytes.NewReader(someBytes[0:127]), 127)

	commPA, err := GeneratePieceCommitmentFromFile(pieceFileA, 127)
	require.NoError(t, err)

	// seek back to head (generating piece commitment moves offset)
	_, err = pieceFileA.Seek(0, 0)
	require.NoError(t, err)

	// write the first piece using the alignment-free function
	n, commP, err := WriteWithoutAlignment(pieceFileA, 127, stagedSectorFile)
	require.NoError(t, err)
	require.Equal(t, int(n), 127)
	require.Equal(t, commP, commPA)

	// write second piece + alignment
	require.NoError(t, err)
	pieceFileB := requireTempFile(t, bytes.NewReader(someBytes[0:508]), 508)

	commPB, err := GeneratePieceCommitmentFromFile(pieceFileB, 508)
	require.NoError(t, err)

	// seek back to head
	_, err = pieceFileB.Seek(0, 0)
	require.NoError(t, err)

	// second piece relies on the alignment-computing version
	left, tot, commP, err := WriteWithAlignment(pieceFileB, 508, stagedSectorFile, []uint64{127})
	require.NoError(t, err)
	require.Equal(t, int(left), 381)
	require.Equal(t, int(tot), 889)
	require.Equal(t, commP, commPB)

	publicPieces := []PublicPieceInfo{{
		Size:  127,
		CommP: commPA,
	}, {
		Size:  508,
		CommP: commPB,
	}}

	commD, err := GenerateDataCommitment(sectorSize, publicPieces)
	require.NoError(t, err)

	privatePieces := make([]PieceMetadata, len(publicPieces))
	for i, v := range publicPieces {
		privatePieces[i] = PieceMetadata{
			Key:   hex.EncodeToString(v.CommP[:]),
			Size:  v.Size,
			CommP: v.CommP,
		}
	}

	// pre-commit the sector
	output, err := SealPreCommit(sectorSize, poRepProofPartitions, sectorCacheDirPath, stagedSectorFile.Name(), sealedSectorFile.Name(), sectorID, proverID, ticket.TicketBytes, publicPieces)
	require.NoError(t, err)

	require.Equal(t, output.CommD, commD, "prover and verifier should agree on data commitment")

	// commit the sector
	proof, err := SealCommit(sectorSize, poRepProofPartitions, sectorCacheDirPath, sectorID, proverID, ticket.TicketBytes, seed.TicketBytes, publicPieces, output)
	require.NoError(t, err)

	// verify the 'ole proofy
	isValid, err := VerifySeal(sectorSize, output.CommR, output.CommD, proverID, ticket.TicketBytes, seed.TicketBytes, sectorID, proof)
	require.NoError(t, err)
	require.True(t, isValid, "proof wasn't valid")

	// unseal the entire sector and verify that things went as we planned
	require.NoError(t, Unseal(sectorSize, poRepProofPartitions, sectorCacheDirPath, sealedSectorFile.Name(), unsealOutputFileA.Name(), sectorID, proverID, ticket.TicketBytes, output.CommD))
	contents, err := ioutil.ReadFile(unsealOutputFileA.Name())
	require.NoError(t, err)

	// unsealed sector includes a bunch of alignment NUL-bytes
	alignment := make([]byte, 381)

	// verify that we unsealed what we expected to unseal
	require.Equal(t, someBytes[0:127], contents[0:127])
	require.Equal(t, alignment, contents[127:508])
	require.Equal(t, someBytes[0:508], contents[508:1016])

	// unseal just the first piece
	err = UnsealRange(sectorSize, poRepProofPartitions, sectorCacheDirPath, sealedSectorFile.Name(), unsealOutputFileB.Name(), sectorID, proverID, ticket.TicketBytes, output.CommD, 0, 127)
	require.NoError(t, err)
	contentsB, err := ioutil.ReadFile(unsealOutputFileB.Name())
	require.NoError(t, err)
	require.Equal(t, 127, len(contentsB))
	require.Equal(t, someBytes[0:127], contentsB[0:127])

	// unseal just the second piece
	err = UnsealRange(sectorSize, poRepProofPartitions, sectorCacheDirPath, sealedSectorFile.Name(), unsealOutputFileC.Name(), sectorID, proverID, ticket.TicketBytes, output.CommD, 508, 508)
	require.NoError(t, err)
	contentsC, err := ioutil.ReadFile(unsealOutputFileC.Name())
	require.NoError(t, err)
	require.Equal(t, 508, len(contentsC))
	require.Equal(t, someBytes[0:508], contentsC[0:508])

	// verify that the sector builder owns no sealed sectors
	var sealedSectorPaths []string
	require.NoError(t, filepath.Walk(sealedSectorsDir, visit(&sealedSectorPaths)))
	assert.Equal(t, 1, len(sealedSectorPaths), sealedSectorPaths)

	// no sector cache dirs, either
	var sectorCacheDirPaths []string
	require.NoError(t, filepath.Walk(sectorCacheRootDir, visit(&sectorCacheDirPaths)))
	assert.Equal(t, 1, len(sectorCacheDirPaths), sectorCacheDirPaths)

	// generate a PoSt over the proving set before importing, just to exercise
	// the new API
	privateInfo := NewSortedPrivateSectorInfo(PrivateSectorInfo{
		SectorID:         sectorID,
		CommR:            output.CommR,
		CacheDirPath:     sectorCacheDirPath,
		SealedSectorPath: sealedSectorFile.Name(),
	})

	publicInfo := NewSortedPublicSectorInfo(PublicSectorInfo{
		SectorID: sectorID,
		CommR:    output.CommR,
	})

	candidatesA, err := GenerateCandidates(sectorSize, proverID, randomness, challengeCount, privateInfo)
	require.NoError(t, err)

	// finalize the ticket, but don't do anything with the results (simply
	// exercise the API)
	_, err = FinalizeTicket(candidatesA[0].PartialTicket)
	require.NoError(t, err)

	proofA, err := GeneratePoSt(sectorSize, proverID, privateInfo, randomness, candidatesA)
	require.NoError(t, err)

	isValid, err = VerifyPoSt(sectorSize, publicInfo, randomness, challengeCount, proofA, candidatesA, proverID)
	require.NoError(t, err)
	require.True(t, isValid, "VerifyPoSt rejected the (standalone) proof as invalid")
}

func TestJsonMarshalSymmetry(t *testing.T) {
	for i := 0; i < 100; i++ {
		xs := make([]PublicSectorInfo, 10)
		for j := 0; j < 10; j++ {
			var x PublicSectorInfo
			_, err := io.ReadFull(rand.Reader, x.CommR[:])
			require.NoError(t, err)

			n, err := rand.Int(rand.Reader, big.NewInt(500))
			require.NoError(t, err)
			x.SectorID = n.Uint64()
			xs[j] = x
		}
		toSerialize := NewSortedPublicSectorInfo(xs...)

		serialized, err := toSerialize.MarshalJSON()
		require.NoError(t, err)

		var fromSerialized SortedPublicSectorInfo
		err = fromSerialized.UnmarshalJSON(serialized)
		require.NoError(t, err)

		require.Equal(t, toSerialize, fromSerialized)
	}
}

func requireTempFile(t *testing.T, fileContentsReader io.Reader, size uint64) *os.File {
	file, err := ioutil.TempFile("", "")
	require.NoError(t, err)

	written, err := io.Copy(file, fileContentsReader)
	require.NoError(t, err)
	// check that we wrote everything
	require.Equal(t, int(size), int(written))

	require.NoError(t, file.Sync())

	// seek to the beginning
	_, err = file.Seek(0, 0)
	require.NoError(t, err)

	return file
}

func requireTempDirPath(t *testing.T, prefix string) string {
	dir, err := ioutil.TempDir("", prefix)
	require.NoError(t, err)

	return dir
}

func visit(paths *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		*paths = append(*paths, path)
		return nil
	}
}
