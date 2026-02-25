# Security Guidelines

## Overview

This document outlines security best practices and the security improvements made to the SDK.

## Security Fixes Implemented

### 1. Secure gRPC Connections

**Issue**: The SDK was using insecure gRPC connections without TLS encryption.

**Fix**: 
- Added `SecureConnection()` function that uses TLS 1.2+ encryption by default
- Deprecated `DefaultConnection()` with warning about insecure usage
- `DefaultConnection()` now returns an error instead of using `log.Fatalf()`

**Usage**:
```go
// RECOMMENDED: Use SecureConnection for production
conn, err := keystone.SecureConnection(host, port, vendorID, appID, accessToken)
if err != nil {
    // Handle error
}

// DEPRECATED: Only use for testing/development
conn, err := keystone.DefaultConnection(host, port, vendorID, appID, accessToken)
```

### 2. Cryptographic Hash Improvements

**Issue**: Test models were using SHA-1, which is cryptographically broken.

**Fix**: Replaced SHA-1 with SHA-256 in test models.

**Example**:
```go
// Before: hash := sha1.Sum([]byte(data))
// After:  hash := sha256.Sum256([]byte(data))
```

### 3. SSRF (Server-Side Request Forgery) Protection

**Issue**: URL fetching in `objects.go` did not validate URLs, allowing potential SSRF attacks.

**Fix**: 
- Added `validateURL()` function that checks URL scheme (only http/https allowed)
- Added URL validation in `NewUploadFromURL()`, `CopyFromURL()`, and `getRemoteFile()`

**Protection**:
- Only HTTP and HTTPS schemes are allowed
- Malformed URLs are rejected
- Prevents access to file://, ftp://, and other dangerous schemes

### 4. Path Traversal Protection

**Issue**: File paths were not validated, allowing potential directory traversal attacks.

**Fix**:
- Added `validatePath()` function that cleans and validates paths
- Checks for ".." sequences after path cleaning
- Applied to `NewUpload()` and `NewUploadFromURL()`

### 5. Panic Handling

**Issue**: Several functions used `panic()` for invalid input, which could cause DoS.

**Fix**:
- Changed `HashID()` and `HashCID()` to return errors instead of panicking
- Added `ValidateEntityState()` helper function for state validation
- Updated `ByHashID()` to include panic warning in documentation
- Updated all callers to handle errors properly

**Example**:
```go
// Before: id := keystone.HashID("#invalid")  // Would panic
// After:
id, err := keystone.HashID("#invalid")
if err != nil {
    // Handle error: ErrHashIDContainsInvalidChar
}

// For entity state validation:
if err := keystone.ValidateEntityState(state); err != nil {
    // Handle error: ErrInvalidEntityState
}
```

### 6. HTTP Client Security

**Issue**: Code was using `http.DefaultClient` without timeouts or security settings.

**Fix**:
- Created `secureHTTPClient()` that returns a configured client with:
  - 30-second timeout
  - Connection pooling limits
  - Idle connection timeout
- Applied to all HTTP operations

### 7. Error Handling

**Issue**: Some errors were silently ignored (e.g., `io.ReadAll()`).

**Fix**: Proper error handling throughout, especially in:
- `getRemoteFile()`: Now properly handles and formats errors
- All HTTP operations: Errors are checked and returned

## Security Best Practices

### For SDK Users

1. **Always use `SecureConnection()` in production**
   ```go
   conn, err := keystone.SecureConnection(host, port, vendorID, appID, token)
   ```

2. **Validate user input before passing to SDK functions**
   ```go
   // Validate entity state before using
   if err := keystone.ValidateEntityState(userProvidedState); err != nil {
       return err
   }
   
   // Use HashID safely
   id, err := keystone.HashID(userInput)
   if err != nil {
       return err
   }
   ```

3. **Handle all errors returned by SDK functions**
   ```go
   obj, err := keystone.NewUpload(path, storageClass)
   if err != nil {
       // Handle error - don't ignore
   }
   ```

4. **Sanitize URLs before passing to `NewUploadFromURL()`**
   - The SDK validates schemes, but additional validation is recommended
   - Consider implementing an allowlist of trusted domains

5. **Be cautious with file paths**
   - Even though the SDK validates paths, ensure your application doesn't construct paths from untrusted input

### For SDK Developers

1. **Never use `panic()` for invalid input**
   - Return errors instead
   - Reserve panics for truly exceptional situations

2. **Always validate external input**
   - URLs should be parsed and validated
   - File paths should be cleaned and checked
   - User-provided strings should be validated

3. **Use secure defaults**
   - TLS should be enabled by default
   - Timeouts should always be set on network operations
   - Use strong cryptographic algorithms (SHA-256+, TLS 1.2+)

4. **Handle errors explicitly**
   - Never use blank identifiers (`_`) to ignore errors
   - Return errors to callers
   - Provide context in error messages

## Reporting Security Issues

If you discover a security vulnerability, please email security@keystonedb.com instead of using the public issue tracker.

## Security Audit History

- **2026-02-25**: Comprehensive security audit and fixes
  - Fixed insecure gRPC connections
  - Improved cryptographic practices
  - Added SSRF and path traversal protection
  - Improved error handling and panic management
