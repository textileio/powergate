use std::mem;
use std::slice::from_raw_parts;
use std::sync::Once;

use ffi_toolkit::{
    c_str_to_pbuf, catch_panic_response, raw_ptr, rust_str_to_c_str, FCPResponseStatus,
};
use filecoin_proofs as api_fns;
use filecoin_proofs::{
    types as api_types, PaddedBytesAmount, PieceInfo, PoRepConfig, PoRepProofPartitions,
    PoStConfig, SectorClass, SectorSize, UnpaddedByteIndex, UnpaddedBytesAmount,
};
use libc;
use storage_proofs::sector::SectorId;

use super::helpers::{
    bls_12_fr_into_bytes, c_to_rust_candidates, c_to_rust_proofs, to_private_replica_info_map,
};
use super::types::*;

/// TODO: document
///
#[no_mangle]
#[cfg(not(target_os = "windows"))]
pub unsafe extern "C" fn write_with_alignment(
    src_fd: libc::c_int,
    src_size: u64,
    dst_fd: libc::c_int,
    existing_piece_sizes_ptr: *const u64,
    existing_piece_sizes_len: libc::size_t,
) -> *mut WriteWithAlignmentResponse {
    catch_panic_response(|| {
        init_log();

        info!("write_with_alignment: start");

        let mut response = WriteWithAlignmentResponse::default();

        let piece_sizes: Vec<UnpaddedBytesAmount> =
            from_raw_parts(existing_piece_sizes_ptr, existing_piece_sizes_len)
                .iter()
                .map(|n| UnpaddedBytesAmount(*n))
                .collect();

        let n = UnpaddedBytesAmount(src_size);

        match api_fns::add_piece(
            FileDescriptorRef::new(src_fd),
            FileDescriptorRef::new(dst_fd),
            n,
            &piece_sizes,
        ) {
            Ok((aligned_bytes_written, comm_p)) => {
                response.comm_p = comm_p;
                response.left_alignment_unpadded = (aligned_bytes_written - n).into();
                response.status_code = FCPResponseStatus::FCPNoError;
                response.total_write_unpadded = aligned_bytes_written.into();
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        }

        info!("write_with_alignment: finish");

        raw_ptr(response)
    })
}

/// TODO: document
///
#[no_mangle]
#[cfg(not(target_os = "windows"))]
pub unsafe extern "C" fn write_without_alignment(
    src_fd: libc::c_int,
    src_size: u64,
    dst_fd: libc::c_int,
) -> *mut WriteWithoutAlignmentResponse {
    catch_panic_response(|| {
        init_log();

        info!("write_without_alignment: start");

        let mut response = WriteWithoutAlignmentResponse::default();

        match api_fns::write_and_preprocess(
            FileDescriptorRef::new(src_fd),
            FileDescriptorRef::new(dst_fd),
            UnpaddedBytesAmount(src_size),
        ) {
            Ok((total_bytes_written, comm_p)) => {
                response.comm_p = comm_p;
                response.status_code = FCPResponseStatus::FCPNoError;
                response.total_write_unpadded = total_bytes_written.into();
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        }

        info!("write_without_alignment: finish");

        raw_ptr(response)
    })
}

/// TODO: document
///
#[no_mangle]
pub unsafe extern "C" fn seal_pre_commit(
    sector_class: FFISectorClass,
    cache_dir_path: *const libc::c_char,
    staged_sector_path: *const libc::c_char,
    sealed_sector_path: *const libc::c_char,
    sector_id: u64,
    prover_id: &[u8; 32],
    ticket: &[u8; 32],
    pieces_ptr: *const FFIPublicPieceInfo,
    pieces_len: libc::size_t,
) -> *mut SealPreCommitResponse {
    catch_panic_response(|| {
        init_log();

        info!("seal_pre_commit: start");

        let mut response = SealPreCommitResponse::default();

        let public_pieces: Vec<PieceInfo> = from_raw_parts(pieces_ptr, pieces_len)
            .iter()
            .cloned()
            .map(Into::into)
            .collect();

        let sc: SectorClass = sector_class.into();

        match api_fns::seal_pre_commit(
            sc.into(),
            c_str_to_pbuf(cache_dir_path),
            c_str_to_pbuf(staged_sector_path),
            c_str_to_pbuf(sealed_sector_path),
            *prover_id,
            SectorId::from(sector_id),
            *ticket,
            &public_pieces,
        ) {
            Ok(output) => {
                response.status_code = FCPResponseStatus::FCPNoError;

                let mut x: FFISealPreCommitOutput = Default::default();
                x.comm_r = output.comm_r;
                x.comm_d = output.comm_d;

                response.seal_pre_commit_output = x;
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        }

        info!("seal_pre_commit: finish");

        raw_ptr(response)
    })
}

/// TODO: document
///
#[no_mangle]
pub unsafe extern "C" fn seal_commit(
    sector_class: FFISectorClass,
    cache_dir_path: *const libc::c_char,
    sector_id: u64,
    prover_id: &[u8; 32],
    ticket: &[u8; 32],
    seed: &[u8; 32],
    pieces_ptr: *const FFIPublicPieceInfo,
    pieces_len: libc::size_t,
    spco: FFISealPreCommitOutput,
) -> *mut SealCommitResponse {
    catch_panic_response(|| {
        init_log();

        info!("seal_commit: start");

        let mut response = SealCommitResponse::default();

        let spco = api_types::SealPreCommitOutput {
            comm_r: spco.comm_r,
            comm_d: spco.comm_d,
        };

        let public_pieces: Vec<PieceInfo> = from_raw_parts(pieces_ptr, pieces_len)
            .iter()
            .cloned()
            .map(Into::into)
            .collect();

        let sc: SectorClass = sector_class.into();

        match api_fns::seal_commit(
            sc.into(),
            c_str_to_pbuf(cache_dir_path),
            *prover_id,
            SectorId::from(sector_id),
            *ticket,
            *seed,
            spco,
            &public_pieces,
        ) {
            Ok(output) => {
                response.status_code = FCPResponseStatus::FCPNoError;
                response.proof_ptr = output.proof.as_ptr();
                response.proof_len = output.proof.len();
                mem::forget(output.proof);
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        }

        info!("seal_commit: finish");

        raw_ptr(response)
    })
}

/// TODO: document
///
#[no_mangle]
pub unsafe extern "C" fn unseal(
    sector_class: FFISectorClass,
    cache_dir_path: *const libc::c_char,
    sealed_sector_path: *const libc::c_char,
    unseal_output_path: *const libc::c_char,
    sector_id: u64,
    prover_id: &[u8; 32],
    ticket: &[u8; 32],
    comm_d: &[u8; 32],
) -> *mut UnsealResponse {
    catch_panic_response(|| {
        init_log();

        info!("unseal: start");

        let sc: SectorClass = sector_class.clone().into();

        let result = api_fns::get_unsealed_range(
            sc.into(),
            c_str_to_pbuf(cache_dir_path),
            c_str_to_pbuf(sealed_sector_path),
            c_str_to_pbuf(unseal_output_path),
            *prover_id,
            SectorId::from(sector_id),
            *comm_d,
            *ticket,
            UnpaddedByteIndex(0u64),
            UnpaddedBytesAmount::from(PaddedBytesAmount(sector_class.sector_size)),
        );

        let mut response = UnsealResponse::default();

        match result {
            Ok(_) => {
                response.status_code = FCPResponseStatus::FCPNoError;
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        };

        info!("unseal: finish");

        raw_ptr(response)
    })
}

/// TODO: document
///
#[no_mangle]
pub unsafe extern "C" fn unseal_range(
    sector_class: FFISectorClass,
    cache_dir_path: *const libc::c_char,
    sealed_sector_path: *const libc::c_char,
    unseal_output_path: *const libc::c_char,
    sector_id: u64,
    prover_id: &[u8; 32],
    ticket: &[u8; 32],
    comm_d: &[u8; 32],
    offset: u64,
    length: u64,
) -> *mut UnsealRangeResponse {
    catch_panic_response(|| {
        init_log();

        info!("unseal_range: start");

        let sc: SectorClass = sector_class.clone().into();

        let result = api_fns::get_unsealed_range(
            sc.into(),
            c_str_to_pbuf(cache_dir_path),
            c_str_to_pbuf(sealed_sector_path),
            c_str_to_pbuf(unseal_output_path),
            *prover_id,
            SectorId::from(sector_id),
            *comm_d,
            *ticket,
            UnpaddedByteIndex(offset),
            UnpaddedBytesAmount(length),
        );

        let mut response = UnsealRangeResponse::default();

        match result {
            Ok(_) => {
                response.status_code = FCPResponseStatus::FCPNoError;
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        };

        info!("unseal_range: finish");

        raw_ptr(response)
    })
}

/// Verifies the output of seal.
///
#[no_mangle]
pub unsafe extern "C" fn verify_seal(
    sector_size: u64,
    comm_r: &[u8; 32],
    comm_d: &[u8; 32],
    prover_id: &[u8; 32],
    ticket: &[u8; 32],
    seed: &[u8; 32],
    sector_id: u64,
    proof_ptr: *const u8,
    proof_len: libc::size_t,
) -> *mut super::types::VerifySealResponse {
    catch_panic_response(|| {
        init_log();

        info!("verify_seal: start");

        let porep_bytes = super::helpers::try_into_porep_proof_bytes(proof_ptr, proof_len);

        let result = porep_bytes.and_then(|bs| {
            super::helpers::porep_proof_partitions_try_from_bytes(&bs).and_then(|partitions| {
                let cfg = api_types::PoRepConfig {
                    sector_size: api_types::SectorSize(sector_size),
                    partitions,
                };

                api_fns::verify_seal(
                    cfg,
                    *comm_r,
                    *comm_d,
                    *prover_id,
                    SectorId::from(sector_id),
                    *ticket,
                    *seed,
                    &bs,
                )
            })
        });

        let mut response = VerifySealResponse::default();

        match result {
            Ok(true) => {
                response.status_code = FCPResponseStatus::FCPNoError;
                response.is_valid = true;
            }
            Ok(false) => {
                response.status_code = FCPResponseStatus::FCPNoError;
                response.is_valid = false;
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        };

        info!("verify_seal: finish");

        raw_ptr(response)
    })
}

/// Finalize a partial_ticket.
#[no_mangle]
pub unsafe extern "C" fn finalize_ticket(partial_ticket: &[u8; 32]) -> *mut FinalizeTicketResponse {
    catch_panic_response(|| {
        init_log();

        info!("finalize_ticket: start");

        let mut response = FinalizeTicketResponse::default();

        match filecoin_proofs::finalize_ticket(partial_ticket) {
            Ok(ticket) => {
                response.status_code = FCPResponseStatus::FCPNoError;
                response.ticket = ticket;
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        };

        info!("finalize_ticket: finish");

        raw_ptr(response)
    })
}

/// Verifies that a proof-of-spacetime is valid.
#[no_mangle]
pub unsafe extern "C" fn verify_post(
    sector_size: u64,
    randomness: &[u8; 32],
    challenge_count: u64,
    sector_ids_ptr: *const u64,
    sector_ids_len: libc::size_t,
    flattened_comm_rs_ptr: *const u8,
    flattened_comm_rs_len: libc::size_t,
    flattened_proofs_ptr: *const u8,
    flattened_proofs_len: libc::size_t,
    winners_ptr: *const FFICandidate,
    winners_len: libc::size_t,
    prover_id: &[u8; 32],
) -> *mut VerifyPoStResponse {
    catch_panic_response(|| {
        init_log();

        info!("verify_post: start");

        let mut response = VerifyPoStResponse::default();

        let convert = super::helpers::to_public_replica_info_map(
            sector_ids_ptr,
            sector_ids_len,
            flattened_comm_rs_ptr,
            flattened_comm_rs_len,
        );

        let result = convert.and_then(|map| {
            let proofs = c_to_rust_proofs(flattened_proofs_ptr, flattened_proofs_len)?;
            let winners = c_to_rust_candidates(winners_ptr, winners_len)?;
            let cfg = api_types::PoStConfig {
                sector_size: api_types::SectorSize(sector_size),
            };
            api_fns::verify_post(
                cfg,
                randomness,
                challenge_count,
                &proofs,
                &map,
                &winners,
                *prover_id,
            )
        });

        match result {
            Ok(is_valid) => {
                response.status_code = FCPResponseStatus::FCPNoError;
                response.is_valid = is_valid;
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        };

        info!("verify_post: {}", "finish");
        raw_ptr(response)
    })
}

/// Returns the merkle root for a piece after piece padding and alignment.
/// The caller is responsible for closing the passed in file descriptor.
#[no_mangle]
#[cfg(not(target_os = "windows"))]
pub unsafe extern "C" fn generate_piece_commitment(
    piece_fd_raw: libc::c_int,
    unpadded_piece_size: u64,
) -> *mut GeneratePieceCommitmentResponse {
    catch_panic_response(|| {
        init_log();

        use std::os::unix::io::{FromRawFd, IntoRawFd};

        let mut piece_file = std::fs::File::from_raw_fd(piece_fd_raw);

        let unpadded_piece_size = api_types::UnpaddedBytesAmount(unpadded_piece_size);

        let result = api_fns::generate_piece_commitment(&mut piece_file, unpadded_piece_size);

        // avoid dropping the File which closes it
        let _ = piece_file.into_raw_fd();

        let mut response = GeneratePieceCommitmentResponse::default();

        match result {
            Ok(meta) => {
                response.status_code = FCPResponseStatus::FCPNoError;
                response.comm_p = meta.commitment;
                response.num_bytes_aligned = meta.size.into();
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        }

        raw_ptr(response)
    })
}

/// Returns the merkle root for a sector containing the provided pieces.
#[no_mangle]
pub unsafe extern "C" fn generate_data_commitment(
    sector_size: u64,
    pieces_ptr: *const FFIPublicPieceInfo,
    pieces_len: libc::size_t,
) -> *mut GenerateDataCommitmentResponse {
    catch_panic_response(|| {
        init_log();

        let public_pieces: Vec<PieceInfo> = from_raw_parts(pieces_ptr, pieces_len)
            .iter()
            .cloned()
            .map(Into::into)
            .collect();

        let result = api_fns::compute_comm_d(
            PoRepConfig {
                sector_size: SectorSize(sector_size),
                partitions: PoRepProofPartitions(0),
            },
            &public_pieces,
        );

        let mut response = GenerateDataCommitmentResponse::default();

        match result {
            Ok(commitment) => {
                response.status_code = FCPResponseStatus::FCPNoError;
                response.comm_d = commitment;
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        }

        raw_ptr(response)
    })
}

/// TODO: document
///
#[no_mangle]
pub unsafe extern "C" fn generate_candidates(
    sector_size: u64,
    randomness: &[u8; 32],
    challenge_count: u64,
    replicas_ptr: *const FFIPrivateReplicaInfo,
    replicas_len: libc::size_t,
    prover_id: &[u8; 32],
) -> *mut GenerateCandidatesResponse {
    catch_panic_response(|| {
        init_log();

        info!("generate_candidates: start");

        let mut response = GenerateCandidatesResponse::default();

        let result = to_private_replica_info_map(replicas_ptr, replicas_len).and_then(|rs| {
            api_fns::generate_candidates(
                PoStConfig {
                    sector_size: SectorSize(sector_size),
                },
                randomness,
                challenge_count,
                &rs,
                *prover_id,
            )
        });

        match result {
            Ok(output) => {
                let mapped: Vec<FFICandidate> = output
                    .iter()
                    .map(|x| FFICandidate {
                        sector_id: x.sector_id.into(),
                        partial_ticket: bls_12_fr_into_bytes(x.partial_ticket),
                        ticket: x.ticket,
                        sector_challenge_index: x.sector_challenge_index,
                    })
                    .collect();

                response.status_code = FCPResponseStatus::FCPNoError;
                response.candidates_ptr = mapped.as_ptr();
                response.candidates_len = mapped.len();
                mem::forget(mapped);
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        }

        info!("generate_candidates: finish");

        raw_ptr(response)
    })
}

/// TODO: document
///
#[no_mangle]
pub unsafe extern "C" fn generate_post(
    sector_size: u64,
    randomness: &[u8; 32],
    replicas_ptr: *const FFIPrivateReplicaInfo,
    replicas_len: libc::size_t,
    winners_ptr: *const FFICandidate,
    winners_len: libc::size_t,
    prover_id: &[u8; 32],
) -> *mut GeneratePoStResponse {
    catch_panic_response(|| {
        init_log();

        info!("generate_post: start");

        let mut response = GeneratePoStResponse::default();

        let result = to_private_replica_info_map(replicas_ptr, replicas_len).and_then(|rs| {
            api_fns::generate_post(
                PoStConfig {
                    sector_size: SectorSize(sector_size),
                },
                randomness,
                &rs,
                c_to_rust_candidates(winners_ptr, winners_len)?,
                *prover_id,
            )
        });

        match result {
            Ok(proof) => {
                response.status_code = FCPResponseStatus::FCPNoError;

                let flattened_proofs: Vec<u8> = proof.into_iter().flatten().collect();

                response.flattened_proofs_len = flattened_proofs.len();
                response.flattened_proofs_ptr = flattened_proofs.as_ptr();

                mem::forget(flattened_proofs);
            }
            Err(err) => {
                response.status_code = FCPResponseStatus::FCPUnclassifiedError;
                response.error_msg = rust_str_to_c_str(format!("{:?}", err));
            }
        }

        info!("generate_post: finish");

        raw_ptr(response)
    })
}

#[no_mangle]
pub unsafe extern "C" fn destroy_write_with_alignment_response(
    ptr: *mut WriteWithAlignmentResponse,
) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn destroy_write_without_alignment_response(
    ptr: *mut WriteWithoutAlignmentResponse,
) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn destroy_seal_pre_commit_response(ptr: *mut SealPreCommitResponse) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn destroy_seal_commit_response(ptr: *mut SealCommitResponse) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn destroy_unseal_response(ptr: *mut UnsealResponse) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn destroy_unseal_range_response(ptr: *mut UnsealRangeResponse) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn destroy_generate_piece_commitment_response(
    ptr: *mut GeneratePieceCommitmentResponse,
) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn destroy_generate_data_commitment_response(
    ptr: *mut GenerateDataCommitmentResponse,
) {
    let _ = Box::from_raw(ptr);
}

/// Returns the number of user bytes that will fit into a staged sector.
///
#[no_mangle]
pub unsafe extern "C" fn get_max_user_bytes_per_staged_sector(sector_size: u64) -> u64 {
    u64::from(api_types::UnpaddedBytesAmount::from(api_types::SectorSize(
        sector_size,
    )))
}

/// Deallocates a VerifySealResponse.
///
#[no_mangle]
pub unsafe extern "C" fn destroy_verify_seal_response(ptr: *mut VerifySealResponse) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn destroy_finalize_ticket_response(ptr: *mut FinalizeTicketResponse) {
    let _ = Box::from_raw(ptr);
}

/// Deallocates a VerifyPoStResponse.
///
#[no_mangle]
pub unsafe extern "C" fn destroy_verify_post_response(ptr: *mut VerifyPoStResponse) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn destroy_generate_post_response(ptr: *mut GeneratePoStResponse) {
    let _ = Box::from_raw(ptr);
}

#[no_mangle]
pub unsafe extern "C" fn destroy_generate_candidates_response(
    ptr: *mut GenerateCandidatesResponse,
) {
    let _ = Box::from_raw(ptr);
}

/// Protects the init off the logger.
static LOG_INIT: Once = Once::new();

/// Ensures the logger is initialized.
fn init_log() {
    LOG_INIT.call_once(|| {
        fil_logger::init();
    });
}

#[cfg(test)]
pub mod tests {
    use std::io::{Read, Seek, SeekFrom, Write};
    use std::os::unix::io::IntoRawFd;

    use anyhow::Result;
    use ffi_toolkit::{c_str_to_rust_str, FCPResponseStatus};
    use rand::{thread_rng, Rng};

    use super::*;

    #[test]
    fn test_write_with_and_without_alignment() -> Result<()> {
        // write some bytes to a temp file to be used as the byte source
        let mut rng = thread_rng();
        let buf: Vec<u8> = (0..508).map(|_| rng.gen()).collect();

        // first temp file occupies 4 nodes in a merkle tree built over the
        // destination (after preprocessing)
        let mut src_file_a = tempfile::tempfile()?;
        let _ = src_file_a.write_all(&buf[0..127])?;
        src_file_a.seek(SeekFrom::Start(0))?;

        // second occupies 16 nodes
        let mut src_file_b = tempfile::tempfile()?;
        let _ = src_file_b.write_all(&buf[0..508])?;
        src_file_b.seek(SeekFrom::Start(0))?;

        // create a temp file to be used as the byte destination
        let dest = tempfile::tempfile()?;

        // transmute temp files to file descriptors
        let src_fd_a = src_file_a.into_raw_fd();
        let src_fd_b = src_file_b.into_raw_fd();
        let dst_fd = dest.into_raw_fd();

        // write the first file
        unsafe {
            let resp = write_without_alignment(src_fd_a, 127, dst_fd);

            if (*resp).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp).error_msg);
                panic!("write_without_alignment failed: {:?}", msg);
            }

            assert_eq!(
                (*resp).total_write_unpadded,
                127,
                "should have added 127 bytes of (unpadded) left alignment"
            );
        }

        // write the second
        unsafe {
            let existing = vec![127u64];

            let resp =
                write_with_alignment(src_fd_b, 508, dst_fd, existing.as_ptr(), existing.len());

            if (*resp).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp).error_msg);
                panic!("write_with_alignment failed: {:?}", msg);
            }

            assert_eq!(
                (*resp).left_alignment_unpadded,
                381,
                "should have added 381 bytes of (unpadded) left alignment"
            );
        }

        Ok(())
    }

    #[test]
    fn test_sealing() -> Result<()> {
        // miscellaneous setup and shared values
        let sector_class = FFISectorClass {
            sector_size: 1024,
            porep_proof_partitions: 2,
        };

        let cache_dir = tempfile::tempdir()?;
        let cache_dir_path = cache_dir.into_path();

        let challenge_count = 1;
        let prover_id = [1u8; 32];
        let randomness = [7u8; 32];
        let sector_id = 42;
        let sector_size = 1024;
        let seed = [5u8; 32];
        let ticket = [6u8; 32];

        // create a byte source (a user's piece)
        let mut rng = thread_rng();
        let buf_a: Vec<u8> = (0..1016).map(|_| rng.gen()).collect();
        let mut piece_file = tempfile::tempfile()?;
        let _ = piece_file.write_all(&buf_a)?;
        piece_file.seek(SeekFrom::Start(0))?;

        // create the staged sector (the byte destination)
        let (staged_file, staged_path) = tempfile::NamedTempFile::new()?.keep()?;

        // create a temp file to be used as the byte destination
        let (_, sealed_path) = tempfile::NamedTempFile::new()?.keep()?;

        // last temp file is used to output unsealed bytes
        let (_, unseal_path) = tempfile::NamedTempFile::new()?.keep()?;

        // transmute temp files to file descriptors
        let piece_file_fd = piece_file.into_raw_fd();
        let staged_sector_fd = staged_file.into_raw_fd();

        unsafe {
            let resp_a = write_without_alignment(piece_file_fd, 1016, staged_sector_fd);

            if (*resp_a).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp_a).error_msg);
                panic!("write_without_alignment failed: {:?}", msg);
            }

            let pieces = vec![FFIPublicPieceInfo {
                num_bytes: 1016,
                comm_p: (*resp_a).comm_p,
            }];

            let cache_dir_path_c_str = rust_str_to_c_str(cache_dir_path.to_str().unwrap());
            let staged_path_c_str = rust_str_to_c_str(staged_path.to_str().unwrap());
            let sealed_path_c_str = rust_str_to_c_str(sealed_path.to_str().unwrap());
            let unseal_path_c_str = rust_str_to_c_str(unseal_path.to_str().unwrap());

            let resp_b = seal_pre_commit(
                sector_class.clone(),
                cache_dir_path_c_str,
                staged_path_c_str,
                sealed_path_c_str,
                sector_id,
                &prover_id,
                &ticket,
                pieces.as_ptr(),
                pieces.len(),
            );

            if (*resp_b).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp_b).error_msg);
                panic!("seal_pre_commit failed: {:?}", msg);
            }

            let resp_c = seal_commit(
                sector_class.clone(),
                cache_dir_path_c_str,
                sector_id,
                &prover_id,
                &ticket,
                &seed,
                pieces.as_ptr(),
                pieces.len(),
                (*(resp_b)).seal_pre_commit_output,
            );

            if (*resp_c).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp_c).error_msg);
                panic!("seal_commit failed: {:?}", msg);
            }

            let resp_d = verify_seal(
                sector_size,
                &(*resp_b).seal_pre_commit_output.comm_r,
                &(*resp_b).seal_pre_commit_output.comm_d,
                &prover_id,
                &ticket,
                &seed,
                sector_id,
                (*resp_c).proof_ptr,
                (*resp_c).proof_len,
            );

            if (*resp_d).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp_d).error_msg);
                panic!("seal_commit failed: {:?}", msg);
            }

            assert!((*resp_d).is_valid, "proof was not valid");

            let resp_e = unseal(
                sector_class.clone(),
                cache_dir_path_c_str,
                sealed_path_c_str,
                unseal_path_c_str,
                sector_id,
                &prover_id,
                &ticket,
                &(*resp_b).seal_pre_commit_output.comm_d,
            );

            if (*resp_e).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp_e).error_msg);
                panic!("unseal failed: {:?}", msg);
            }

            // ensure unsealed bytes match what we had in our piece
            let mut buf_b = Vec::with_capacity(1016);
            let mut f = std::fs::File::open(unseal_path)?;
            let _ = f.read_to_end(&mut buf_b)?;
            assert_eq!(
                format!("{:x?}", &buf_a),
                format!("{:x?}", &buf_b),
                "original bytes don't match unsealed bytes"
            );

            // generate a PoSt

            let private_replicas = vec![FFIPrivateReplicaInfo {
                cache_dir_path: cache_dir_path_c_str,
                comm_r: (*resp_b).seal_pre_commit_output.comm_r,
                replica_path: sealed_path_c_str,
                sector_id,
            }];

            let resp_f = generate_candidates(
                sector_size,
                &randomness,
                challenge_count,
                private_replicas.as_ptr(),
                private_replicas.len(),
                &prover_id,
            );

            if (*resp_f).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp_f).error_msg);
                panic!("generate_candidates failed: {:?}", msg);
            }

            // exercise the ticket-finalizing code path (but don't do anything
            // with the results
            let result = c_to_rust_candidates((*resp_f).candidates_ptr, (*resp_f).candidates_len)?;
            if result.len() < 1 {
                panic!("generate_candidates produced no results");
            }

            let resp_g = finalize_ticket(&bls_12_fr_into_bytes(result[0].partial_ticket));
            if (*resp_g).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp_g).error_msg);
                panic!("finalize_ticket failed: {:?}", msg);
            }

            let resp_h = generate_post(
                sector_size,
                &randomness,
                private_replicas.as_ptr(),
                private_replicas.len(),
                (*resp_f).candidates_ptr,
                (*resp_f).candidates_len,
                &prover_id,
            );

            if (*resp_h).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp_h).error_msg);
                panic!("generate_post failed: {:?}", msg);
            }

            let resp_i = verify_post(
                sector_size,
                &randomness,
                challenge_count,
                &sector_id as *const u64,
                1,
                &(*resp_b).seal_pre_commit_output.comm_r[0],
                32,
                (*resp_h).flattened_proofs_ptr,
                (*resp_h).flattened_proofs_len,
                (*resp_f).candidates_ptr,
                (*resp_f).candidates_len,
                &prover_id,
            );

            if (*resp_i).status_code != FCPResponseStatus::FCPNoError {
                let msg = c_str_to_rust_str((*resp_i).error_msg);
                panic!("verify_post failed: {:?}", msg);
            }

            if !(*resp_i).is_valid {
                panic!("verify_post rejected the provided proof as invalid");
            }

            destroy_write_without_alignment_response(resp_a);
            destroy_seal_pre_commit_response(resp_b);
            destroy_seal_commit_response(resp_c);
            destroy_verify_seal_response(resp_d);
            destroy_unseal_response(resp_e);
            destroy_generate_candidates_response(resp_f);
            destroy_finalize_ticket_response(resp_g);
            destroy_generate_post_response(resp_h);
            destroy_verify_post_response(resp_i);

            c_str_to_rust_str(cache_dir_path_c_str);
            c_str_to_rust_str(staged_path_c_str);
            c_str_to_rust_str(sealed_path_c_str);
            c_str_to_rust_str(unseal_path_c_str);
        }

        Ok(())
    }
}
