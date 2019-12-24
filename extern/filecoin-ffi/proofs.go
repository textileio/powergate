package ffi

import (
	"bytes"
	"encoding/json"
	"os"
	"runtime"
	"sort"
	"unsafe"

	"github.com/pkg/errors"
)

// #cgo LDFLAGS: ${SRCDIR}/libfilecoin.a
// #cgo pkg-config: ${SRCDIR}/filecoin.pc
// #include "./filecoin.h"
import "C"

// SortedPublicSectorInfo is a slice of PublicSectorInfo sorted
// (lexicographically, ascending) by replica commitment (CommR).
type SortedPublicSectorInfo struct {
	f []PublicSectorInfo
}

// SortedPrivateSectorInfo is a slice of PrivateSectorInfo sorted
// (lexicographically, ascending) by replica commitment (CommR).
type SortedPrivateSectorInfo struct {
	f []PrivateSectorInfo
}

// SealTicket is required for the first step of Interactive PoRep.
type SealTicket struct {
	BlockHeight uint64
	TicketBytes [32]byte
}

// SealSeed is required for the second step of Interactive PoRep.
type SealSeed struct {
	BlockHeight uint64
	TicketBytes [32]byte
}

type Candidate struct {
	SectorID             uint64
	PartialTicket        [32]byte
	Ticket               [32]byte
	SectorChallengeIndex uint64
}

// NewSortedPublicSectorInfo returns a SortedPublicSectorInfo
func NewSortedPublicSectorInfo(sectorInfo ...PublicSectorInfo) SortedPublicSectorInfo {
	fn := func(i, j int) bool {
		return bytes.Compare(sectorInfo[i].CommR[:], sectorInfo[j].CommR[:]) == -1
	}

	sort.Slice(sectorInfo[:], fn)

	return SortedPublicSectorInfo{
		f: sectorInfo,
	}
}

// Values returns the sorted PublicSectorInfo as a slice
func (s *SortedPublicSectorInfo) Values() []PublicSectorInfo {
	return s.f
}

// MarshalJSON JSON-encodes and serializes the SortedPublicSectorInfo.
func (s SortedPublicSectorInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.f)
}

// UnmarshalJSON parses the JSON-encoded byte slice and stores the result in the
// value pointed to by s.f. Note that this method allows for construction of a
// SortedPublicSectorInfo which violates its invariant (that its PublicSectorInfo are sorted
// in some defined way). Callers should take care to never provide a byte slice
// which would violate this invariant.
func (s *SortedPublicSectorInfo) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.f)
}

type PublicSectorInfo struct {
	SectorID uint64
	CommR    [CommitmentBytesLen]byte
}

// NewSortedPrivateSectorInfo returns a SortedPrivateSectorInfo
func NewSortedPrivateSectorInfo(sectorInfo ...PrivateSectorInfo) SortedPrivateSectorInfo {
	fn := func(i, j int) bool {
		return bytes.Compare(sectorInfo[i].CommR[:], sectorInfo[j].CommR[:]) == -1
	}

	sort.Slice(sectorInfo[:], fn)

	return SortedPrivateSectorInfo{
		f: sectorInfo,
	}
}

// Values returns the sorted PrivateSectorInfo as a slice
func (s *SortedPrivateSectorInfo) Values() []PrivateSectorInfo {
	return s.f
}

// MarshalJSON JSON-encodes and serializes the SortedPrivateSectorInfo.
func (s SortedPrivateSectorInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.f)
}

func (s *SortedPrivateSectorInfo) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.f)
}

type PrivateSectorInfo struct {
	SectorID         uint64
	CommR            [CommitmentBytesLen]byte
	CacheDirPath     string
	SealedSectorPath string
}

// CommitmentBytesLen is the number of bytes in a CommR, CommD, CommP, and CommRStar.
const CommitmentBytesLen = 32

// SealPreCommitOutput is used to acquire a seed from the chain for the second
// step of Interactive PoRep.
type SealPreCommitOutput struct {
	SectorID uint64
	CommD    [CommitmentBytesLen]byte
	CommR    [CommitmentBytesLen]byte
	Pieces   []PieceMetadata
	Ticket   SealTicket
}

// RawSealPreCommitOutput is used to acquire a seed from the chain for the
// second step of Interactive PoRep.
type RawSealPreCommitOutput struct {
	CommD     [CommitmentBytesLen]byte
	CommR     [CommitmentBytesLen]byte
}

// SealCommitOutput is produced by the second step of Interactive PoRep.
type SealCommitOutput struct {
	SectorID uint64
	CommD    [CommitmentBytesLen]byte
	CommR    [CommitmentBytesLen]byte
	Proof    []byte
	Pieces   []PieceMetadata
	Ticket   SealTicket
	Seed     SealSeed
}

// PieceMetadata represents a piece stored by the sector builder.
type PieceMetadata struct {
	Key   string
	Size  uint64
	CommP [CommitmentBytesLen]byte
}

// PublicPieceInfo is an on-chain tuple of CommP and aligned piece-size.
type PublicPieceInfo struct {
	Size  uint64
	CommP [CommitmentBytesLen]byte
}

// VerifySeal returns true if the sealing operation from which its inputs were
// derived was valid, and false if not.
func VerifySeal(
	sectorSize uint64,
	commR [CommitmentBytesLen]byte,
	commD [CommitmentBytesLen]byte,
	proverID [32]byte,
	ticket [32]byte,
	seed [32]byte,
	sectorID uint64,
	proof []byte,
) (bool, error) {

	commDCBytes := C.CBytes(commD[:])
	defer C.free(commDCBytes)

	commRCBytes := C.CBytes(commR[:])
	defer C.free(commRCBytes)

	proofCBytes := C.CBytes(proof[:])
	defer C.free(proofCBytes)

	proverIDCBytes := C.CBytes(proverID[:])
	defer C.free(proverIDCBytes)

	ticketCBytes := C.CBytes(ticket[:])
	defer C.free(ticketCBytes)

	seedCBytes := C.CBytes(seed[:])
	defer C.free(seedCBytes)

	// a mutable pointer to a VerifySealResponse C-struct
	resPtr := C.verify_seal(
		C.uint64_t(sectorSize),
		(*[CommitmentBytesLen]C.uint8_t)(commRCBytes),
		(*[CommitmentBytesLen]C.uint8_t)(commDCBytes),
		(*[32]C.uint8_t)(proverIDCBytes),
		(*[32]C.uint8_t)(ticketCBytes),
		(*[32]C.uint8_t)(seedCBytes),
		C.uint64_t(sectorID),
		(*C.uint8_t)(proofCBytes),
		C.size_t(len(proof)),
	)
	defer C.destroy_verify_seal_response(resPtr)

	if resPtr.status_code != 0 {
		return false, errors.New(C.GoString(resPtr.error_msg))
	}

	return bool(resPtr.is_valid), nil
}

// VerifyPoSt returns true if the PoSt-generation operation from which its
// inputs were derived was valid, and false if not.
func VerifyPoSt(
	sectorSize uint64,
	sectorInfo SortedPublicSectorInfo,
	randomness [32]byte,
	challengeCount uint64,
	proof []byte,
	winners []Candidate,
	proverID [32]byte,
) (bool, error) {
	// CommRs and sector ids must be provided to C.verify_post in the same order
	// that they were provided to the C.generate_post
	sortedCommRs := make([][CommitmentBytesLen]byte, len(sectorInfo.Values()))
	sortedSectorIds := make([]uint64, len(sectorInfo.Values()))
	for idx, v := range sectorInfo.Values() {
		sortedCommRs[idx] = v.CommR
		sortedSectorIds[idx] = v.SectorID
	}

	// flattening the byte slice makes it easier to copy into the C heap
	flattened := make([]byte, CommitmentBytesLen*len(sortedCommRs))
	for idx, commR := range sortedCommRs {
		copy(flattened[(CommitmentBytesLen*idx):(CommitmentBytesLen*(1+idx))], commR[:])
	}

	// copy bytes from Go to C heap
	flattenedCommRsCBytes := C.CBytes(flattened)
	defer C.free(flattenedCommRsCBytes)

	randomnessCBytes := C.CBytes(randomness[:])
	defer C.free(randomnessCBytes)

	proofCBytes := C.CBytes(proof)
	defer C.free(proofCBytes)

	// allocate fixed-length array of uint64s in C heap
	sectorIdsPtr, sectorIdsSize := cUint64s(sortedSectorIds)
	defer C.free(unsafe.Pointer(sectorIdsPtr))

	winnersPtr, winnersSize := cCandidates(winners)
	defer C.free(unsafe.Pointer(winnersPtr))

	proverIDCBytes := C.CBytes(proverID[:])
	defer C.free(proverIDCBytes)

	// a mutable pointer to a VerifyPoStResponse C-struct
	resPtr := C.verify_post(
		C.uint64_t(sectorSize),
		(*[32]C.uint8_t)(randomnessCBytes),
		C.uint64_t(challengeCount),
		sectorIdsPtr,
		sectorIdsSize,
		(*C.uint8_t)(flattenedCommRsCBytes),
		C.size_t(len(flattened)),
		(*C.uint8_t)(proofCBytes),
		C.size_t(len(proof)),
		winnersPtr,
		winnersSize,
		(*[32]C.uint8_t)(proverIDCBytes),
	)
	defer C.destroy_verify_post_response(resPtr)

	if resPtr.status_code != 0 {
		return false, errors.New(C.GoString(resPtr.error_msg))
	}

	return bool(resPtr.is_valid), nil
}

// GetMaxUserBytesPerStagedSector returns the number of user bytes that will fit
// into a staged sector. Due to bit-padding, the number of user bytes that will
// fit into the staged sector will be less than number of bytes in sectorSize.
func GetMaxUserBytesPerStagedSector(sectorSize uint64) uint64 {
	return uint64(C.get_max_user_bytes_per_staged_sector(C.uint64_t(sectorSize)))
}

// GeneratePieceCommitment produces a piece commitment for the provided data
// stored at a given path.
func GeneratePieceCommitment(piecePath string, pieceSize uint64) ([CommitmentBytesLen]byte, error) {
	pieceFile, err := os.Open(piecePath)
	if err != nil {
		return [CommitmentBytesLen]byte{}, err
	}

	return GeneratePieceCommitmentFromFile(pieceFile, pieceSize)
}

// GenerateDataCommitment produces a commitment for the sector containing the
// provided pieces.
func GenerateDataCommitment(sectorSize uint64, pieces []PublicPieceInfo) ([CommitmentBytesLen]byte, error) {
	cPiecesPtr, cPiecesLen := cPublicPieceInfo(pieces)
	defer C.free(unsafe.Pointer(cPiecesPtr))

	resPtr := C.generate_data_commitment(C.uint64_t(sectorSize), (*C.FFIPublicPieceInfo)(cPiecesPtr), cPiecesLen)
	defer C.destroy_generate_data_commitment_response(resPtr)

	if resPtr.status_code != 0 {
		return [CommitmentBytesLen]byte{}, errors.New(C.GoString(resPtr.error_msg))
	}

	return goCommitment(&resPtr.comm_d[0]), nil
}

// GeneratePieceCommitmentFromFile produces a piece commitment for the provided data
// stored in a given file.
func GeneratePieceCommitmentFromFile(pieceFile *os.File, pieceSize uint64) (commP [CommitmentBytesLen]byte, err error) {
	pieceFd := pieceFile.Fd()

	resPtr := C.generate_piece_commitment(C.int(pieceFd), C.uint64_t(pieceSize))
	defer C.destroy_generate_piece_commitment_response(resPtr)

	// Make sure our filedescriptor stays alive, stayin alive
	runtime.KeepAlive(pieceFile)

	if resPtr.status_code != 0 {
		return [CommitmentBytesLen]byte{}, errors.New(C.GoString(resPtr.error_msg))
	}

	return goCommitment(&resPtr.comm_p[0]), nil
}

// WriteWithAlignment
func WriteWithAlignment(
	pieceFile *os.File,
	pieceBytes uint64,
	stagedSectorFile *os.File,
	existingPieceSizes []uint64,
) (leftAlignment, total uint64, commP [CommitmentBytesLen]byte, retErr error) {
	pieceFd := pieceFile.Fd()
	runtime.KeepAlive(pieceFile)

	stagedSectorFd := stagedSectorFile.Fd()
	runtime.KeepAlive(stagedSectorFile)

	ptr, len := cUint64s(existingPieceSizes)
	defer C.free(unsafe.Pointer(ptr))

	resPtr := C.write_with_alignment(
		C.int(pieceFd),
		C.uint64_t(pieceBytes),
		C.int(stagedSectorFd),
		ptr,
		len,
	)
	defer C.destroy_write_with_alignment_response(resPtr)

	if resPtr.status_code != 0 {
		return 0, 0, [CommitmentBytesLen]byte{}, errors.New(C.GoString(resPtr.error_msg))
	}

	return uint64(resPtr.left_alignment_unpadded), uint64(resPtr.total_write_unpadded), goCommitment(&resPtr.comm_p[0]), nil
}

// WriteWithoutAlignment
func WriteWithoutAlignment(
	pieceFile *os.File,
	pieceBytes uint64,
	stagedSectorFile *os.File,
) (uint64, [CommitmentBytesLen]byte, error) {
	pieceFd := pieceFile.Fd()
	runtime.KeepAlive(pieceFile)

	stagedSectorFd := stagedSectorFile.Fd()
	runtime.KeepAlive(stagedSectorFile)

	resPtr := C.write_without_alignment(
		C.int(pieceFd),
		C.uint64_t(pieceBytes),
		C.int(stagedSectorFd),
	)
	defer C.destroy_write_without_alignment_response(resPtr)

	if resPtr.status_code != 0 {
		return 0, [CommitmentBytesLen]byte{}, errors.New(C.GoString(resPtr.error_msg))
	}

	return uint64(resPtr.total_write_unpadded), goCommitment(&resPtr.comm_p[0]), nil
}

// SealPreCommit
func SealPreCommit(
	sectorSize uint64,
	poRepProofPartitions uint8,
	cacheDirPath string,
	stagedSectorPath string,
	sealedSectorPath string,
	sectorID uint64,
	proverID [32]byte,
	ticket [32]byte,
	pieces []PublicPieceInfo,
) (RawSealPreCommitOutput, error) {
	cCacheDirPath := C.CString(cacheDirPath)
	defer C.free(unsafe.Pointer(cCacheDirPath))

	cStagedSectorPath := C.CString(stagedSectorPath)
	defer C.free(unsafe.Pointer(cStagedSectorPath))

	cSealedSectorPath := C.CString(sealedSectorPath)
	defer C.free(unsafe.Pointer(cSealedSectorPath))

	proverIDCBytes := C.CBytes(proverID[:])
	defer C.free(proverIDCBytes)

	ticketCBytes := C.CBytes(ticket[:])
	defer C.free(ticketCBytes)

	cPiecesPtr, cPiecesLen := cPublicPieceInfo(pieces)
	defer C.free(unsafe.Pointer(cPiecesPtr))

	resPtr := C.seal_pre_commit(
		cSectorClass(sectorSize, poRepProofPartitions),
		cCacheDirPath,
		cStagedSectorPath,
		cSealedSectorPath,
		C.uint64_t(sectorID),
		(*[32]C.uint8_t)(proverIDCBytes),
		(*[32]C.uint8_t)(ticketCBytes),
		(*C.FFIPublicPieceInfo)(cPiecesPtr),
		cPiecesLen,
	)
	defer C.destroy_seal_pre_commit_response(resPtr)

	if resPtr.status_code != 0 {
		return RawSealPreCommitOutput{}, errors.New(C.GoString(resPtr.error_msg))
	}

	return goRawSealPreCommitOutput(resPtr.seal_pre_commit_output), nil
}

// SealCommit
func SealCommit(
	sectorSize uint64,
	poRepProofPartitions uint8,
	cacheDirPath string,
	sectorID uint64,
	proverID [32]byte,
	ticket [32]byte,
	seed [32]byte,
	pieces []PublicPieceInfo,
	rspco RawSealPreCommitOutput,
) ([]byte, error) {
	cCacheDirPath := C.CString(cacheDirPath)
	defer C.free(unsafe.Pointer(cCacheDirPath))

	proverIDCBytes := C.CBytes(proverID[:])
	defer C.free(proverIDCBytes)

	ticketCBytes := C.CBytes(ticket[:])
	defer C.free(ticketCBytes)

	seedCBytes := C.CBytes(seed[:])
	defer C.free(seedCBytes)

	cPiecesPtr, cPiecesLen := cPublicPieceInfo(pieces)
	defer C.free(unsafe.Pointer(cPiecesPtr))

	resPtr := C.seal_commit(
		cSectorClass(sectorSize, poRepProofPartitions),
		cCacheDirPath,
		C.uint64_t(sectorID),
		(*[32]C.uint8_t)(proverIDCBytes),
		(*[32]C.uint8_t)(ticketCBytes),
		(*[32]C.uint8_t)(seedCBytes),
		(*C.FFIPublicPieceInfo)(cPiecesPtr),
		cPiecesLen,
		cSealPreCommitOutput(rspco),
	)
	defer C.destroy_seal_commit_response(resPtr)

	if resPtr.status_code != 0 {
		return nil, errors.New(C.GoString(resPtr.error_msg))
	}

	return C.GoBytes(unsafe.Pointer(resPtr.proof_ptr), C.int(resPtr.proof_len)), nil
}

// Unseal
func Unseal(
	sectorSize uint64,
	poRepProofPartitions uint8,
	cacheDirPath string,
	sealedSectorPath string,
	unsealOutputPath string,
	sectorID uint64,
	proverID [32]byte,
	ticket [32]byte,
	commD [CommitmentBytesLen]byte,
) error {
	cCacheDirPath := C.CString(cacheDirPath)
	defer C.free(unsafe.Pointer(cCacheDirPath))

	cSealedSectorPath := C.CString(sealedSectorPath)
	defer C.free(unsafe.Pointer(cSealedSectorPath))

	cUnsealOutputPath := C.CString(unsealOutputPath)
	defer C.free(unsafe.Pointer(cUnsealOutputPath))

	proverIDCBytes := C.CBytes(proverID[:])
	defer C.free(proverIDCBytes)

	ticketCBytes := C.CBytes(ticket[:])
	defer C.free(ticketCBytes)

	commDCBytes := C.CBytes(commD[:])
	defer C.free(commDCBytes)

	resPtr := C.unseal(
		cSectorClass(sectorSize, poRepProofPartitions),
		cCacheDirPath,
		cSealedSectorPath,
		cUnsealOutputPath,
		C.uint64_t(sectorID),
		(*[32]C.uint8_t)(proverIDCBytes),
		(*[32]C.uint8_t)(ticketCBytes),
		(*[CommitmentBytesLen]C.uint8_t)(commDCBytes),
	)
	defer C.destroy_unseal_response(resPtr)

	if resPtr.status_code != 0 {
		return errors.New(C.GoString(resPtr.error_msg))
	}

	return nil
}


// UnsealRange
func UnsealRange(
	sectorSize uint64,
	poRepProofPartitions uint8,
	cacheDirPath string,
	sealedSectorPath string,
	unsealOutputPath string,
	sectorID uint64,
	proverID [32]byte,
	ticket [32]byte,
	commD [CommitmentBytesLen]byte,
	offset uint64,
	len uint64,
) error {
	cCacheDirPath := C.CString(cacheDirPath)
	defer C.free(unsafe.Pointer(cCacheDirPath))

	cSealedSectorPath := C.CString(sealedSectorPath)
	defer C.free(unsafe.Pointer(cSealedSectorPath))

	cUnsealOutputPath := C.CString(unsealOutputPath)
	defer C.free(unsafe.Pointer(cUnsealOutputPath))

	proverIDCBytes := C.CBytes(proverID[:])
	defer C.free(proverIDCBytes)

	ticketCBytes := C.CBytes(ticket[:])
	defer C.free(ticketCBytes)

	commDCBytes := C.CBytes(commD[:])
	defer C.free(commDCBytes)

	resPtr := C.unseal_range(
		cSectorClass(sectorSize, poRepProofPartitions),
		cCacheDirPath,
		cSealedSectorPath,
		cUnsealOutputPath,
		C.uint64_t(sectorID),
		(*[32]C.uint8_t)(proverIDCBytes),
		(*[32]C.uint8_t)(ticketCBytes),
		(*[CommitmentBytesLen]C.uint8_t)(commDCBytes),
		C.uint64_t(offset),
		C.uint64_t(len),
	)
	defer C.destroy_unseal_range_response(resPtr)

	if resPtr.status_code != 0 {
		return errors.New(C.GoString(resPtr.error_msg))
	}

	return nil
}

// FinalizeTicket creates an actual ticket from a partial ticket.
func FinalizeTicket(partialTicket [32]byte) ([32]byte, error) {
	partialTicketPtr := unsafe.Pointer(&(partialTicket)[0])
	resPtr := C.finalize_ticket(
		(*[32]C.uint8_t)(partialTicketPtr),
	)
	defer C.destroy_finalize_ticket_response(resPtr)

	if resPtr.status_code != 0 {
		return [32]byte{}, errors.New(C.GoString(resPtr.error_msg))
	}

	return goCommitment(&resPtr.ticket[0]), nil
}

// GenerateCandidates
func GenerateCandidates(
	sectorSize uint64,
	proverID [32]byte,
	randomness [32]byte,
	challengeCount uint64,
	privateSectorInfo SortedPrivateSectorInfo,
) ([]Candidate, error) {
	randomessCBytes := C.CBytes(randomness[:])
	defer C.free(randomessCBytes)

	proverIDCBytes := C.CBytes(proverID[:])
	defer C.free(proverIDCBytes)

	replicasPtr, replicasSize := cPrivateReplicaInfos(privateSectorInfo.Values())
	defer C.free(unsafe.Pointer(replicasPtr))

	resPtr := C.generate_candidates(
		C.uint64_t(sectorSize),
		(*[32]C.uint8_t)(randomessCBytes),
		C.uint64_t(challengeCount),
		replicasPtr,
		replicasSize,
		(*[32]C.uint8_t)(proverIDCBytes),
	)
	defer C.destroy_generate_candidates_response(resPtr)

	if resPtr.status_code != 0 {
		return nil, errors.New(C.GoString(resPtr.error_msg))
	}

	return goCandidates(resPtr.candidates_ptr, resPtr.candidates_len)
}

// GeneratePoSt
func GeneratePoSt(
	sectorSize uint64,
	proverID [32]byte,
	privateSectorInfo SortedPrivateSectorInfo,
	randomness [32]byte,
	winners []Candidate,
) ([]byte, error) {
	replicasPtr, replicasSize := cPrivateReplicaInfos(privateSectorInfo.Values())
	defer C.free(unsafe.Pointer(replicasPtr))

	winnersPtr, winnersSize := cCandidates(winners)
	defer C.free(unsafe.Pointer(winnersPtr))

	proverIDCBytes := C.CBytes(proverID[:])
	defer C.free(proverIDCBytes)

	resPtr := C.generate_post(
		C.uint64_t(sectorSize),
		(*[32]C.uint8_t)(unsafe.Pointer(&(randomness)[0])),
		replicasPtr,
		replicasSize,
		winnersPtr,
		winnersSize,
		(*[32]C.uint8_t)(proverIDCBytes),
	)
	defer C.destroy_generate_post_response(resPtr)

	if resPtr.status_code != 0 {
		return nil, errors.New(C.GoString(resPtr.error_msg))
	}

	return goBytes(resPtr.flattened_proofs_ptr, resPtr.flattened_proofs_len), nil
}

// SingleProofPartitionProofLen denotes the number of bytes in a proof generated
// with a single partition. The number of bytes in a proof increases linearly
// with the number of partitions used when creating that proof.
const SingleProofPartitionProofLen = 192

func cPublicPieceInfo(src []PublicPieceInfo) (*C.FFIPublicPieceInfo, C.size_t) {
	srcCSizeT := C.size_t(len(src))

	// allocate array in C heap
	cPublicPieceInfos := C.malloc(srcCSizeT * C.sizeof_FFIPublicPieceInfo)

	// create a Go slice backed by the C-array
	xs := (*[1 << 30]C.FFIPublicPieceInfo)(cPublicPieceInfos)
	for i, v := range src {
		xs[i] = C.FFIPublicPieceInfo{
			num_bytes: C.uint64_t(v.Size),
			comm_p:    *(*[32]C.uint8_t)(unsafe.Pointer(&v.CommP)),
		}
	}

	return (*C.FFIPublicPieceInfo)(cPublicPieceInfos), srcCSizeT
}

func cUint64s(src []uint64) (*C.uint64_t, C.size_t) {
	srcCSizeT := C.size_t(len(src))

	// allocate array in C heap
	cUint64s := C.malloc(srcCSizeT * C.sizeof_uint64_t)

	// create a Go slice backed by the C-array
	pp := (*[1 << 30]C.uint64_t)(cUint64s)
	for i, v := range src {
		pp[i] = C.uint64_t(v)
	}

	return (*C.uint64_t)(cUint64s), srcCSizeT
}

func cSectorClass(sectorSize uint64, poRepProofPartitions uint8) C.FFISectorClass {
	return C.FFISectorClass{
		sector_size:            C.uint64_t(sectorSize),
		porep_proof_partitions: C.uint8_t(poRepProofPartitions),
	}
}

func cSealPreCommitOutput(src RawSealPreCommitOutput) C.FFISealPreCommitOutput {
	return C.FFISealPreCommitOutput{
		comm_d:            *(*[32]C.uint8_t)(unsafe.Pointer(&src.CommD)),
		comm_r:            *(*[32]C.uint8_t)(unsafe.Pointer(&src.CommR)),
	}
}

func cCandidates(src []Candidate) (*C.FFICandidate, C.size_t) {
	srcCSizeT := C.size_t(len(src))

	// allocate array in C heap
	cCandidates := C.malloc(srcCSizeT * C.sizeof_FFICandidate)

	// create a Go slice backed by the C-array
	pp := (*[1 << 30]C.FFICandidate)(cCandidates)
	for i, v := range src {
		pp[i] = C.FFICandidate{
			sector_id:              C.uint64_t(v.SectorID),
			partial_ticket:         *(*[32]C.uint8_t)(unsafe.Pointer(&v.PartialTicket)),
			ticket:                 *(*[32]C.uint8_t)(unsafe.Pointer(&v.Ticket)),
			sector_challenge_index: C.uint64_t(v.SectorChallengeIndex),
		}
	}

	return (*C.FFICandidate)(cCandidates), srcCSizeT
}

func cPrivateReplicaInfos(src []PrivateSectorInfo) (*C.FFIPrivateReplicaInfo, C.size_t) {
	srcCSizeT := C.size_t(len(src))

	cPrivateReplicas := C.malloc(srcCSizeT * C.sizeof_FFIPrivateReplicaInfo)

	pp := (*[1 << 30]C.FFIPrivateReplicaInfo)(cPrivateReplicas)
	for i, v := range src {
		pp[i] = C.FFIPrivateReplicaInfo{
			cache_dir_path: C.CString(v.CacheDirPath),
			comm_r:         *(*[32]C.uint8_t)(unsafe.Pointer(&v.CommR)),
			replica_path:   C.CString(v.SealedSectorPath),
			sector_id:      C.uint64_t(v.SectorID),
		}
	}

	return (*C.FFIPrivateReplicaInfo)(cPrivateReplicas), srcCSizeT
}

func goBytes(src *C.uint8_t, size C.size_t) []byte {
	return C.GoBytes(unsafe.Pointer(src), C.int(size))
}

func goCandidates(src *C.FFICandidate, size C.size_t) ([]Candidate, error) {
	candidates := make([]Candidate, size)
	if src == nil || size == 0 {
		return candidates, nil
	}

	ptrs := (*[1 << 30]C.FFICandidate)(unsafe.Pointer(src))[:size:size]
	for i := 0; i < int(size); i++ {
		candidates[i] = goCandidate(ptrs[i])
	}

	return candidates, nil
}

func goCandidate(src C.FFICandidate) Candidate {
	return Candidate{
		SectorID:             uint64(src.sector_id),
		PartialTicket:        goCommitment(&src.partial_ticket[0]),
		Ticket:               goCommitment(&src.ticket[0]),
		SectorChallengeIndex: uint64(src.sector_challenge_index),
	}
}

func goRawSealPreCommitOutput(src C.FFISealPreCommitOutput) RawSealPreCommitOutput {
	return RawSealPreCommitOutput{
		CommD:     goCommitment(&src.comm_d[0]),
		CommR:     goCommitment(&src.comm_r[0]),
	}
}

func goCommitment(src *C.uint8_t) [32]byte {
	slice := C.GoBytes(unsafe.Pointer(src), 32)
	var array [CommitmentBytesLen]byte
	copy(array[:], slice)

	return array
}
