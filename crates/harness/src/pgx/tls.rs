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

use std::path::PathBuf;
use std::sync::Arc;

use rustls::client::WebPkiServerVerifier;
use rustls::client::danger::{HandshakeSignatureValid, ServerCertVerified, ServerCertVerifier};
use rustls::pki_types::pem::PemObject;
use rustls::pki_types::{CertificateDer, PrivateKeyDer, ServerName, UnixTime};
use rustls::{
    CertificateError, ClientConfig, DigitallySignedStruct, Error as TlsError, RootCertStore, SignatureScheme,
};

use crate::pgx::error::Error;

/// SslMode is libpq's sslmode as pgx interprets it.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum SslMode {
    Disable,
    Allow,
    Prefer,
    Require,
    VerifyCa,
    VerifyFull,
}

impl SslMode {
    /// parse parses an sslmode value.
    pub fn parse(value: &str) -> Result<SslMode, Error> {
        Ok(match value {
            "disable" => SslMode::Disable,
            "allow" => SslMode::Allow,
            "prefer" => SslMode::Prefer,
            "require" => SslMode::Require,
            "verify-ca" => SslMode::VerifyCa,
            "verify-full" => SslMode::VerifyFull,
            _ => return Err(Error::Other(format!("sslmode is invalid: {value}"))),
        })
    }
}

/// TlsSettings holds the TLS connection parameters.
#[derive(Clone, Debug)]
pub struct TlsSettings {
    pub mode: SslMode,
    pub root_cert: Option<PathBuf>,
    pub cert: Option<PathBuf>,
    pub key: Option<PathBuf>,
}

impl Default for TlsSettings {
    fn default() -> TlsSettings {
        TlsSettings { mode: SslMode::Prefer, root_cert: None, cert: None, key: None }
    }
}

/// Attempt is one connection attempt in the order pgconn tries them.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub(crate) enum Attempt {
    Plain,
    Tls,
}

impl TlsSettings {
    /// attempts returns the connection attempts for the mode, in order.
    pub(crate) fn attempts(&self) -> &'static [Attempt] {
        match self.mode {
            SslMode::Disable => &[Attempt::Plain],
            SslMode::Allow => &[Attempt::Plain, Attempt::Tls],
            SslMode::Prefer => &[Attempt::Tls, Attempt::Plain],
            SslMode::Require | SslMode::VerifyCa | SslMode::VerifyFull => &[Attempt::Tls],
        }
    }

    /// client_config builds the rustls configuration, verifying like pgx: not at all for prefer and allow, and for
    /// require unless a root certificate is given, which then verifies like verify-ca.
    pub(crate) fn client_config(&self) -> Result<Arc<ClientConfig>, Error> {
        let other = |e: std::fmt::Arguments| Error::Other(e.to_string());
        let provider = Arc::new(rustls::crypto::aws_lc_rs::default_provider());
        let builder = ClientConfig::builder_with_provider(provider.clone())
            .with_safe_default_protocol_versions()
            .map_err(|e| other(format_args!("{e}")))?;
        let verify = match self.mode {
            SslMode::VerifyFull => Some(true),
            SslMode::VerifyCa => Some(false),
            SslMode::Require if self.root_cert.is_some() => Some(false),
            _ => None,
        };
        let verifier: Arc<dyn ServerCertVerifier> = match verify {
            None => Arc::new(NoVerification(provider)),
            Some(check_name) => {
                let mut roots = RootCertStore::empty();
                if let Some(path) = &self.root_cert {
                    for cert in CertificateDer::pem_file_iter(path).map_err(|e| other(format_args!("{e}")))? {
                        roots
                            .add(cert.map_err(|e| other(format_args!("{e}")))?)
                            .map_err(|e| other(format_args!("{e}")))?;
                    }
                }
                let inner = WebPkiServerVerifier::builder_with_provider(Arc::new(roots), provider)
                    .build()
                    .map_err(|e| other(format_args!("{e}")))?;
                Arc::new(ChainVerification { inner, check_name })
            }
        };
        let builder = builder.dangerous().with_custom_certificate_verifier(verifier);
        let config = match (&self.cert, &self.key) {
            (Some(cert), Some(key)) => {
                let certs = CertificateDer::pem_file_iter(cert)
                    .map_err(|e| other(format_args!("{e}")))?
                    .collect::<Result<Vec<_>, _>>()
                    .map_err(|e| other(format_args!("{e}")))?;
                let key = PrivateKeyDer::from_pem_file(key).map_err(|e| other(format_args!("{e}")))?;
                builder.with_client_auth_cert(certs, key).map_err(|e| other(format_args!("{e}")))?
            }
            _ => builder.with_no_client_auth(),
        };
        Ok(Arc::new(config))
    }
}

/// server_name returns the TLS server name of a host.
pub(crate) fn server_name(host: &str) -> Result<ServerName<'static>, Error> {
    ServerName::try_from(host.to_string()).map_err(|e| Error::Other(e.to_string()))
}

/// NoVerification accepts any server certificate, like Go's InsecureSkipVerify.
#[derive(Debug)]
struct NoVerification(Arc<rustls::crypto::CryptoProvider>);

impl ServerCertVerifier for NoVerification {
    fn verify_server_cert(
        &self,
        _: &CertificateDer<'_>,
        _: &[CertificateDer<'_>],
        _: &ServerName<'_>,
        _: &[u8],
        _: UnixTime,
    ) -> Result<ServerCertVerified, TlsError> {
        Ok(ServerCertVerified::assertion())
    }

    fn verify_tls12_signature(
        &self,
        message: &[u8],
        cert: &CertificateDer<'_>,
        dss: &DigitallySignedStruct,
    ) -> Result<HandshakeSignatureValid, TlsError> {
        rustls::crypto::verify_tls12_signature(message, cert, dss, &self.0.signature_verification_algorithms)
    }

    fn verify_tls13_signature(
        &self,
        message: &[u8],
        cert: &CertificateDer<'_>,
        dss: &DigitallySignedStruct,
    ) -> Result<HandshakeSignatureValid, TlsError> {
        rustls::crypto::verify_tls13_signature(message, cert, dss, &self.0.signature_verification_algorithms)
    }

    fn supported_verify_schemes(&self) -> Vec<SignatureScheme> {
        self.0.signature_verification_algorithms.supported_schemes()
    }
}

/// ChainVerification verifies the certificate chain, and the host name only when asked to.
#[derive(Debug)]
struct ChainVerification {
    inner: Arc<WebPkiServerVerifier>,
    check_name: bool,
}

impl ServerCertVerifier for ChainVerification {
    fn verify_server_cert(
        &self,
        end_entity: &CertificateDer<'_>,
        intermediates: &[CertificateDer<'_>],
        server_name: &ServerName<'_>,
        ocsp: &[u8],
        now: UnixTime,
    ) -> Result<ServerCertVerified, TlsError> {
        match self.inner.verify_server_cert(end_entity, intermediates, server_name, ocsp, now) {
            Err(TlsError::InvalidCertificate(CertificateError::NotValidForName))
            | Err(TlsError::InvalidCertificate(CertificateError::NotValidForNameContext { .. }))
                if !self.check_name =>
            {
                Ok(ServerCertVerified::assertion())
            }
            result => result,
        }
    }

    fn verify_tls12_signature(
        &self,
        message: &[u8],
        cert: &CertificateDer<'_>,
        dss: &DigitallySignedStruct,
    ) -> Result<HandshakeSignatureValid, TlsError> {
        self.inner.verify_tls12_signature(message, cert, dss)
    }

    fn verify_tls13_signature(
        &self,
        message: &[u8],
        cert: &CertificateDer<'_>,
        dss: &DigitallySignedStruct,
    ) -> Result<HandshakeSignatureValid, TlsError> {
        self.inner.verify_tls13_signature(message, cert, dss)
    }

    fn supported_verify_schemes(&self) -> Vec<SignatureScheme> {
        self.inner.supported_verify_schemes()
    }
}
