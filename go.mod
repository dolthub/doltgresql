module github.com/dolthub/doltgresql

go 1.24.6

require (
	github.com/PuerkitoBio/goquery v1.8.1
	github.com/cockroachdb/apd/v2 v2.0.3-0.20200518165714-d020e156310a
	github.com/cockroachdb/errors v1.7.5
	github.com/dolthub/dolt/go v0.40.5-0.20250917230656-0b4401d4bd16
	github.com/dolthub/eventsapi_schema v0.0.0-20250915094920-eadfd39051ca
	github.com/dolthub/flatbuffers/v23 v23.3.3-dh.2
	github.com/dolthub/go-mysql-server v0.20.1-0.20250917194252-e24900c62d2f
	github.com/dolthub/sqllogictest/go v0.0.0-20240618184124-ca47f9354216
	github.com/dolthub/vitess v0.0.0-20250915221346-753c44800850
	github.com/fatih/color v1.13.0
	github.com/goccy/go-json v0.10.2
	github.com/gogo/protobuf v1.3.2
	github.com/golang/geo v0.0.0-20200730024412-e86565bf3f35
	github.com/google/btree v1.1.2
	github.com/google/go-cmp v0.7.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/jackc/pglogrepl v0.0.0-20240307033717-828fbfe908e9
	github.com/jackc/pgx/v4 v4.18.2
	github.com/jackc/pgx/v5 v5.6.1-0.20240826124046-97d20ccfadaa
	github.com/lib/pq v1.10.9
	github.com/madflojo/testcerts v1.1.1
	github.com/mitchellh/go-ps v1.0.0
	github.com/mitchellh/go-wordwrap v1.0.1
	github.com/pganalyze/pg_query_go/v6 v6.1.0
	github.com/pierrre/geohash v1.0.0
	github.com/pkg/profile v1.5.0
	github.com/sergi/go-diff v1.1.0
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.10.0
	github.com/twpayne/go-geom v1.3.6
	github.com/xdg-go/stringprep v1.0.4
	golang.org/x/crypto v0.40.0
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1
	golang.org/x/net v0.42.0
	golang.org/x/sync v0.16.0
	golang.org/x/sys v0.34.0
	golang.org/x/text v0.27.0
	gopkg.in/src-d/go-errors.v1 v1.0.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go v0.120.0 // indirect
	cloud.google.com/go/auth v0.16.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	cloud.google.com/go/monitoring v1.24.2 // indirect
	cloud.google.com/go/storage v1.50.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.27.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.50.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.50.0 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db // indirect
	github.com/aliyun/aliyun-oss-go-sdk v2.2.5+incompatible // indirect
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/apache/thrift v0.13.1-0.20201008052519-daf620915714 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.8 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.61 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.64 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.41.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.78.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.29.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.16 // indirect
	github.com/aws/smithy-go v1.22.2 // indirect
	github.com/bcicen/jstream v1.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20250501225837-2ac532fd4443 // indirect
	github.com/cockroachdb/logtags v0.0.0-20190617123548-eb05cc24525f // indirect
	github.com/cockroachdb/redact v1.0.6 // indirect
	github.com/cockroachdb/sentry-go v0.6.1-cockroachdb.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/denisbrodbeck/machineid v1.0.1 // indirect
	github.com/dolthub/aws-sdk-go-ini-parser v0.0.0-20250305001723-2821c37f6c12 // indirect
	github.com/dolthub/dolt-mcp v0.2.2-0.20250917171427-13e4520d1c36 // indirect
	github.com/dolthub/fslock v0.0.3 // indirect
	github.com/dolthub/go-icu-regex v0.0.0-20250916051405-78a38d478790 // indirect
	github.com/dolthub/gozstd v0.0.0-20240423170813-23a2903bca63 // indirect
	github.com/dolthub/ishell v0.0.0-20240701202509-2b217167d718 // indirect
	github.com/dolthub/jsonpath v0.0.2-0.20240227200619-19675ab05c71 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/edsrzf/mmap-go v1.2.0 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/esote/minmaxheap v1.0.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/gocraft/dbr/v2 v2.7.2 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-github/v57 v57.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.2 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/juju/gnuflag v0.0.0-20171113085948-2ce1bb71843d // indirect
	github.com/kch42/buzhash v0.0.0-20160816060738-9bdec3dec7c6 // indirect
	github.com/klauspost/compress v1.10.10 // indirect
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lestrrat-go/strftime v1.0.4 // indirect
	github.com/mark3labs/mcp-go v0.34.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mohae/uvarint v0.0.0-20160208145430-c3f9e62bf2b0 // indirect
	github.com/oracle/oci-go-sdk/v65 v65.55.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/silvasur/buzhash v0.0.0-20160816060738-9bdec3dec7c6 // indirect
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966 // indirect
	github.com/sony/gobreaker v0.5.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spiffe/go-spiffe/v2 v2.5.0 // indirect
	github.com/tealeg/xlsx v1.0.5 // indirect
	github.com/tidwall/gjson v1.14.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/twpayne/go-kml v1.5.2-0.20200728095708-9f2fd4dfcbfe // indirect
	github.com/vbauerster/mpb/v8 v8.0.2 // indirect
	github.com/xitongsys/parquet-go v1.6.1 // indirect
	github.com/xitongsys/parquet-go-source v0.0.0-20211010230925-397910c5e371 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.36.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.36.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/sdk v1.36.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/trace v1.36.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/term v0.33.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	golang.org/x/tools v0.34.0 // indirect
	google.golang.org/api v0.241.0 // indirect
	google.golang.org/genproto v0.0.0-20250505200425-f936aa4a68b2 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/grpc v1.74.2 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/pganalyze/pg_query_go/v6 => github.com/dolthub/pg_query_go/v6 v6.0.0-20250702135351-29eb6bfc4ea6
