# Migration Guide for Security Updates

This guide helps you update your code to work with the security improvements in this release.

## Breaking Changes

### 1. DefaultConnection() now returns an error

**Before:**
```go
conn := keystone.DefaultConnection(host, port, vendorID, appID, accessToken)
```

**After:**
```go
conn, err := keystone.DefaultConnection(host, port, vendorID, appID, accessToken)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}

// RECOMMENDED: Use SecureConnection instead for production
conn, err := keystone.SecureConnection(host, port, vendorID, appID, accessToken)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
```

### 2. NewUpload() now returns an error

**Before:**
```go
obj := keystone.NewUpload("path/to/file", proto.ObjectType_Standard)
```

**After:**
```go
obj, err := keystone.NewUpload("path/to/file", proto.ObjectType_Standard)
if err != nil {
    return err
}
```

### 3. HashID() and HashCID() now return errors

**Before:**
```go
id := keystone.HashID("my-id")
childID := keystone.HashCID("parent", "child")
```

**After:**
```go
id, err := keystone.HashID("my-id")
if err != nil {
    return err
}

childID, err := keystone.HashCID("parent", "child")
if err != nil {
    return err
}
```

### 4. SetHashID() error handling

**Before:**
```go
entity.SetHashID(userInput) // Could panic on invalid input
```

**After:**
```go
if err := entity.SetHashID(userInput); err != nil {
    return fmt.Errorf("invalid hash ID: %w", err)
}
```

## Non-Breaking Changes

### 1. New SecureConnection() function

Use this for production deployments:

```go
conn, err := keystone.SecureConnection(host, port, vendorID, appID, accessToken)
if err != nil {
    return fmt.Errorf("failed to connect securely: %w", err)
}
```

### 2. ValidateEntityState() helper

You can now validate entity states before passing them to WithState():

```go
userProvidedState := proto.EntityState_Active // From user input

if err := keystone.ValidateEntityState(userProvidedState); err != nil {
    return fmt.Errorf("invalid state: %w", err)
}

// Safe to use now
err := actor.Mutate(ctx, entity, keystone.WithState(userProvidedState))
```

## Error Codes

The following new error variables are available:

```go
keystone.ErrHashIDContainsInvalidChar  // HashID input contains '#'
keystone.ErrInvalidEntityState         // Invalid or Removed state
keystone.ErrInvalidURL                 // Malformed URL
keystone.ErrUnsupportedScheme          // Non-HTTP(S) URL scheme
keystone.ErrInvalidPath                // Path traversal detected
```

You can check for specific errors:

```go
id, err := keystone.HashID(input)
if errors.Is(err, keystone.ErrHashIDContainsInvalidChar) {
    // Handle specific error
}
```

## Testing Your Migration

1. Update all calls to changed functions
2. Run your tests: `go test ./...`
3. Check for compilation errors
4. Review any panic recovery code - it may no longer be needed

## Finding Affected Code

Use these commands to find code that needs updating:

```bash
# Find DefaultConnection calls
grep -r "DefaultConnection(" --include="*.go"

# Find NewUpload calls
grep -r "NewUpload(" --include="*.go"

# Find HashID/HashCID calls
grep -r "HashID\|HashCID" --include="*.go"

# Find SetHashID calls
grep -r "SetHashID(" --include="*.go"
```

## Need Help?

If you encounter issues during migration:

1. Check the [SECURITY.md](SECURITY.md) documentation
2. Review the test files for examples of proper usage
3. Open an issue on GitHub

## Timeline

- **Deprecated**: `DefaultConnection()` without error return (use new signature)
- **New**: `SecureConnection()` for production use
- **Changed**: Error returns added to several functions

We recommend migrating to the new APIs as soon as possible for improved security.
