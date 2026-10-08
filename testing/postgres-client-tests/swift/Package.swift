// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "postgresnio-test",
    dependencies: [
        .package(url: "https://github.com/vapor/postgres-nio.git", from: "1.19.0"),
    ],
    targets: [
        .executableTarget(
            name: "postgresnio-test",
            dependencies: [
                .product(name: "PostgresNIO", package: "postgres-nio"),
            ],
            path: "Sources"
        ),
    ]
)
