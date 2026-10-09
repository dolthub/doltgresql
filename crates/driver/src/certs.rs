// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

use std::path::Path;

use rcgen::{
    BasicConstraints, CertificateParams, DnType, ExtendedKeyUsagePurpose, IsCa, Issuer, KeyPair, KeyUsagePurpose,
    PKCS_ED25519, PKCS_RSA_SHA256, RsaKeySize, SanType, SerialNumber,
};
use time::{Duration, OffsetDateTime};

const LEAF_DNS: &str = "dolt-instance.dolt-integration-test.example";
const LEAF_URI: &str = "spiffe://dolt-integration-tests.dev.trust.dolthub.com.example/dolt-instance";

/// Chain is a root, an intermediate signed by it, and a valid and an expired leaf signed by the intermediate.
struct Chain {
    root: String,
    intermediate: String,
    leaf: String,
    leaf_key: String,
    expired_leaf: String,
    expired_leaf_key: String,
}

/// params returns certificate parameters with the test subject.
fn params(
    common_name: String,
    serial: u64,
    not_before: OffsetDateTime,
    not_after: OffsetDateTime,
) -> CertificateParams {
    let mut params = CertificateParams::default();
    params.distinguished_name.push(DnType::CountryName, "US");
    params.distinguished_name.push(DnType::OrganizationName, "DoltHub, Inc.");
    params.distinguished_name.push(DnType::CommonName, common_name);
    params.serial_number = Some(SerialNumber::from(serial));
    params.not_before = not_before;
    params.not_after = not_after;
    params
}

/// make_chain generates a chain like the Go suite's MakeCerts, with keys from the generator.
fn make_chain(description: &str, generate: impl Fn() -> Result<KeyPair, rcgen::Error>) -> Result<Chain, rcgen::Error> {
    let not_before = OffsetDateTime::now_utc() - Duration::hours(24);
    let expires = not_before + Duration::hours(24 * 365 * 10);
    let expired = not_before + Duration::hours(12);
    let ca = |name: &str, serial| {
        let mut p = params(format!("dolt integration tests {description} {name}"), serial, not_before, expires);
        p.is_ca = IsCa::Ca(BasicConstraints::Unconstrained);
        p.key_usages = vec![KeyUsagePurpose::KeyCertSign];
        p
    };
    let leaf = |name: &str, serial, not_after| -> Result<CertificateParams, rcgen::Error> {
        let mut p = params(format!("dolt integration tests {description} {name}"), serial, not_before, not_after);
        p.is_ca = IsCa::ExplicitNoCa;
        p.key_usages = vec![KeyUsagePurpose::DigitalSignature];
        p.extended_key_usages = vec![ExtendedKeyUsagePurpose::ServerAuth];
        p.subject_alt_names = vec![SanType::DnsName(LEAF_DNS.try_into()?), SanType::URI(LEAF_URI.try_into()?)];
        Ok(p)
    };
    let root_key = generate()?;
    let root_params = ca("Root", 1);
    let root = root_params.self_signed(&root_key)?;
    let root_issuer = Issuer::from_params(&root_params, &root_key);
    let intermediate_key = generate()?;
    let intermediate_params = ca("Intermediate", 2);
    let intermediate = intermediate_params.signed_by(&intermediate_key, &root_issuer)?;
    let intermediate_issuer = Issuer::from_params(&intermediate_params, &intermediate_key);
    let leaf_key = generate()?;
    let leaf_cert = leaf("Leaf", 3, expires)?.signed_by(&leaf_key, &intermediate_issuer)?;
    let expired_key = generate()?;
    let expired_cert = leaf("Expired Leaf", 4, expired)?.signed_by(&expired_key, &intermediate_issuer)?;
    Ok(Chain {
        root: root.pem(),
        intermediate: intermediate.pem(),
        leaf: leaf_cert.pem(),
        leaf_key: leaf_key.serialize_pem(),
        expired_leaf: expired_cert.pem(),
        expired_leaf_key: expired_key.serialize_pem(),
    })
}

/// generate_x509_certs writes an RSA chain and an ed25519 chain into the directory, with PKCS #8 keys.
pub fn generate_x509_certs(dir: &Path) -> Result<(), String> {
    let rsa = make_chain("rsa", || KeyPair::generate_rsa_for(&PKCS_RSA_SHA256, RsaKeySize::_4096))
        .map_err(|e| format!("could not make rsa certs: {e}"))?;
    let ed = make_chain("ed25519", || KeyPair::generate_for(&PKCS_ED25519))
        .map_err(|e| format!("could not make ed25519 certs: {e}"))?;
    let files = [
        ("rsa_root.pem", rsa.root.clone()),
        ("rsa_chain.pem", format!("{}{}", rsa.leaf, rsa.intermediate)),
        ("rsa_key.pem", rsa.leaf_key),
        ("rsa_exp_chain.pem", format!("{}{}", rsa.expired_leaf, rsa.intermediate)),
        ("rsa_exp_key.pem", rsa.expired_leaf_key),
        ("ed25519_root.pem", ed.root),
        ("ed25519_chain.pem", format!("{}{}", ed.leaf, ed.intermediate)),
        ("ed25519_key.pem", ed.leaf_key),
        ("ed25519_exp_chain.pem", format!("{}{}", ed.expired_leaf, ed.intermediate)),
        ("edcerts_exp_key.pem", ed.expired_leaf_key),
    ];
    for (name, contents) in files {
        std::fs::write(dir.join(name), contents).map_err(|e| format!("{name}: {e}"))?;
    }
    Ok(())
}
