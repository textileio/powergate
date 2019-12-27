use crate::bls::api::{BLSDigest, BLSPrivateKey, BLSPublicKey, BLSSignature};

/// HashResponse

#[repr(C)]
pub struct HashResponse {
    pub digest: BLSDigest,
}

#[no_mangle]
pub unsafe extern "C" fn destroy_hash_response(ptr: *mut HashResponse) {
    let _ = Box::from_raw(ptr);
}

/// AggregateResponse

#[repr(C)]
pub struct AggregateResponse {
    pub signature: BLSSignature,
}

#[no_mangle]
pub unsafe extern "C" fn destroy_aggregate_response(ptr: *mut AggregateResponse) {
    let _ = Box::from_raw(ptr);
}

/// PrivateKeyGenerateResponse

#[repr(C)]
pub struct PrivateKeyGenerateResponse {
    pub private_key: BLSPrivateKey,
}

#[no_mangle]
pub unsafe extern "C" fn destroy_private_key_generate_response(
    ptr: *mut PrivateKeyGenerateResponse,
) {
    let _ = Box::from_raw(ptr);
}

/// PrivateKeySignResponse

#[repr(C)]
pub struct PrivateKeySignResponse {
    pub signature: BLSSignature,
}

#[no_mangle]
pub unsafe extern "C" fn destroy_private_key_sign_response(ptr: *mut PrivateKeySignResponse) {
    let _ = Box::from_raw(ptr);
}

/// PrivateKeyPublicKeyResponse

#[repr(C)]
pub struct PrivateKeyPublicKeyResponse {
    pub public_key: BLSPublicKey,
}

#[no_mangle]
pub unsafe extern "C" fn destroy_private_key_public_key_response(
    ptr: *mut PrivateKeyPublicKeyResponse,
) {
    let _ = Box::from_raw(ptr);
}
