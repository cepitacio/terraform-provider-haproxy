# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-09-13

### Added

-   **Complete HAProxy Management**: Full support for frontends, backends, servers, ACLs, and rules
-   **Data Discovery**: Powerful data sources for discovering existing HAProxy configurations
-   **Atomic Operations**: Stack resource for creating complete HAProxy configurations in single transactions
-   **Multi-Version Support**: Compatible with HAProxy Data Plane API v2 and v3 (v3 recommended)
-   **Comprehensive Coverage**: Support for HTTP/TCP rules, health checks, binds, and advanced features
-   **Concurrent Operations**: Support for multiple `haproxy_stack` resources with `for_each`
-   **Transaction Retry Logic**: Robust error handling with automatic retry for transaction conflicts
-   **Version-Aware Operations**: Different handling for API v2 (individual operations) vs v3 (bulk operations)

### Changed

-   **Major Architecture Overhaul**: Migrated from Terraform SDK v2 to Plugin Framework
-   **API Version Support**: Upgraded from v2-only to v2+v3 support
-   **Resource Structure**: Moved from individual resources to atomic `haproxy_stack` resource
-   **Naming Consistency**: Standardized all manager functions to use "Create" prefix
-   **Schema Organization**: Moved ACLs and rules to proper nested blocks within frontend/backend

### Fixed

-   **Deletion Index Shifting**: Fixed critical bug where deleting items caused subsequent items' indices to shift
-   **Concurrent Operations**: Resolved race conditions when multiple stacks run simultaneously
-   **Transaction Handling**: Fixed "transaction does not exist" and "transaction outdated" errors
-   **API v3 Compatibility**: Fixed 404 errors in v3 by implementing bulk replace operations
-   **Error Accumulation**: Fixed issue where errors persisted across retry attempts
-   **Schema Consistency**: Removed top-level ACLs, properly nested all rules

### Technical Details

-   **Plugin Framework**: Built on Terraform Plugin Framework for better performance and maintainability
-   **Atomic Transactions**: All operations within a stack are committed as a single transaction
-   **Retry Logic**: Up to 3 retries with 2-second delays for transaction conflicts
-   **Error Handling**: Comprehensive error handling with proper diagnostics management
-   **Code Quality**: Consistent naming conventions and clean architecture

## [0.0.8] - Previous Release

### Added

-   Initial release with basic HAProxy management capabilities
-   Support for HAProxy Data Plane API v2 only
-   Individual resource management (frontend, backend, server, etc.)