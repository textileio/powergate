use std::collections::btree_map::BTreeMap;
use std::slice::from_raw_parts;

use anyhow::Result;
use ffi_toolkit::{c_str_to_pbuf, c_str_to_rust_str};
use filecoin_proofs::{constants as api_constants, Commitment, PublicReplicaInfo};
use filecoin_proofs::{types as api_types, PrivateReplicaInfo};
use libc;
use paired::bls12_381::{Bls12, Fr};
use storage_proofs::election_post::Candidate;
use storage_proofs::fr32::fr_into_bytes;
use storage_proofs::sector::SectorId;

use super::types::{FFICandidate, FFIPrivateReplicaInfo};

/// Produce a map from sector id to replica info by pairing sector ids and
/// replica commitments (by index in their respective arrays), setting the
/// storage fault-boolean if the sector id is present in the provided dynamic
/// array. This function's return value should be provided to the verify_post
/// call.
///
pub unsafe fn to_public_replica_info_map(
    sector_ids_ptr: *const u64,
    sector_ids_len: libc::size_t,
    flattened_comm_rs_ptr: *const u8,
    flattened_comm_rs_len: libc::size_t,
) -> Result<BTreeMap<SectorId, PublicReplicaInfo>> {
    ensure!(!sector_ids_ptr.is_null(), "sector_ids_ptr must not be null");
    ensure!(
        !flattened_comm_rs_ptr.is_null(),
        "flattened_comm_rs_ptr must not be null"
    );

    let sector_ids: Vec<SectorId> = from_raw_parts(sector_ids_ptr, sector_ids_len)
        .iter()
        .cloned()
        .map(SectorId::from)
        .collect();

    let comm_rs: Vec<Commitment> = into_commitments(flattened_comm_rs_ptr, flattened_comm_rs_len);

    ensure!(
        sector_ids.len() == comm_rs.len(),
        "must provide equal number of sector ids and replica commitments (sector_ids.len={}, comm_rs.len()={})",
        sector_ids.len(),
        comm_rs.len());

    let mut m = BTreeMap::new();

    for i in 0..sector_ids.len() {
        m.insert(sector_ids[i], PublicReplicaInfo::new(comm_rs[i])?);
    }

    Ok(m)
}

/// Copy the provided dynamic array's bytes into a vector and return the vector.
///
pub unsafe fn try_into_porep_proof_bytes(
    proof_ptr: *const u8,
    proof_len: libc::size_t,
) -> Result<Vec<u8>> {
    into_proof_vecs(proof_len, proof_ptr, proof_len)?
        .first()
        .map(Vec::clone)
        .ok_or_else(|| format_err!("no proofs in chunked vec"))
}

/// Splits the flattened, dynamic array of CommR bytes into a vector of
/// 32-element byte arrays and returns the vector. Each byte array's
/// little-endian value represents an Fr.
///
pub unsafe fn into_commitments(
    flattened_comms_ptr: *const u8,
    flattened_comms_len: libc::size_t,
) -> Vec<[u8; 32]> {
    from_raw_parts(flattened_comms_ptr, flattened_comms_len)
        .iter()
        .step_by(32)
        .fold(Default::default(), |mut acc: Vec<[u8; 32]>, item| {
            let sliced = from_raw_parts(item, 32);
            let mut x: [u8; 32] = Default::default();
            x.copy_from_slice(&sliced[..32]);
            acc.push(x);
            acc
        })
}

/// Return the number of partitions used to create the given proof.
///
pub fn porep_proof_partitions_try_from_bytes(
    proof: &[u8],
) -> Result<api_types::PoRepProofPartitions> {
    let n = proof.len();

    ensure!(
        n % api_constants::SINGLE_PARTITION_PROOF_LEN == 0,
        "no PoRepProofPartitions mapping for {:x?}",
        proof
    );

    Ok(api_types::PoRepProofPartitions(
        (n / api_constants::SINGLE_PARTITION_PROOF_LEN) as u8,
    ))
}

unsafe fn into_proof_vecs(
    proof_chunk: usize,
    flattened_proofs_ptr: *const u8,
    flattened_proofs_len: libc::size_t,
) -> Result<Vec<Vec<u8>>> {
    ensure!(
        !flattened_proofs_ptr.is_null(),
        "flattened_proofs_ptr must not be a null pointer"
    );
    ensure!(proof_chunk > 0, "Invalid proof chunk of 0 passed");

    let res = from_raw_parts(flattened_proofs_ptr, flattened_proofs_len)
        .iter()
        .step_by(proof_chunk)
        .fold(Default::default(), |mut acc: Vec<Vec<u8>>, item| {
            let sliced = from_raw_parts(item, proof_chunk);
            acc.push(sliced.to_vec());
            acc
        });

    Ok(res)
}

pub fn bls_12_fr_into_bytes(fr: Fr) -> [u8; 32] {
    let mut commitment = [0; 32];
    for (i, b) in fr_into_bytes::<Bls12>(&fr).iter().enumerate() {
        commitment[i] = *b;
    }
    commitment
}

pub unsafe fn to_private_replica_info_map(
    replicas_ptr: *const FFIPrivateReplicaInfo,
    replicas_len: libc::size_t,
) -> Result<BTreeMap<SectorId, PrivateReplicaInfo>> {
    ensure!(!replicas_ptr.is_null(), "replicas_ptr must not be null");

    from_raw_parts(replicas_ptr, replicas_len)
        .iter()
        .cloned()
        .map(|ffi_r| {
            let cache_dir_path = c_str_to_pbuf(ffi_r.cache_dir_path);
            let cloned = cache_dir_path.clone();

            filecoin_proofs::PrivateReplicaInfo::new(
                c_str_to_rust_str(ffi_r.replica_path).to_string(),
                ffi_r.comm_r,
                cache_dir_path,
            )
            .map_err(|err| {
                format_err!(
                    "could not load private replica (id = {}) from cache (path = {:?}): {}",
                    ffi_r.sector_id,
                    cloned,
                    err
                )
            })
            .map(|p| (SectorId::from(ffi_r.sector_id), p))
        })
        .collect()
}

pub unsafe fn c_to_rust_candidates(
    winners_ptr: *const FFICandidate,
    winners_len: libc::size_t,
) -> Result<Vec<Candidate>> {
    ensure!(!winners_ptr.is_null(), "winners_ptr must not be null");

    from_raw_parts(winners_ptr, winners_len)
        .iter()
        .cloned()
        .map(|c| c.try_into_candidate().map_err(Into::into))
        .collect()
}

pub unsafe fn c_to_rust_proofs(
    flattened_proofs_ptr: *const u8,
    flattened_proofs_len: libc::size_t,
) -> Result<Vec<Vec<u8>>> {
    ensure!(
        !flattened_proofs_ptr.is_null(),
        "flattened_proof_ptr must not be null"
    );

    Ok(from_raw_parts(flattened_proofs_ptr, flattened_proofs_len)
        .chunks(filecoin_proofs::SINGLE_PARTITION_PROOF_LEN)
        .map(Into::into)
        .collect::<Vec<Vec<u8>>>())
}
