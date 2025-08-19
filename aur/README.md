# Axiom CLI AUR Package Maintenance

This directory contains the files needed to maintain the `axiom-bin` package in the Arch User Repository (AUR).

## Background

The Arch Linux user reported an issue where the AUR package was trying to install the ARM64 version on AMD64 systems. This happened because GoReleaser's AUR integration doesn't properly handle multiple architectures in the generated PKGBUILD file.

### The Problem

GoReleaser generates a single PKGBUILD that downloads a fixed archive URL, which doesn't adapt based on the system architecture (`$CARCH` variable in Arch Linux). This means the package would fail on systems with different architectures than what was hardcoded.

### The Solution

We've implemented a multi-architecture aware PKGBUILD that:
- Uses architecture-specific source arrays (`source_x86_64` and `source_aarch64`)
- Downloads the correct binary based on the system architecture
- Properly validates checksums for each architecture

## Files

- `PKGBUILD` - The multi-architecture PKGBUILD template
- `update-pkgbuild.sh` - Script to update version and checksums after a release
- `README.md` - This documentation

## Release Process

When releasing a new version of axiom-cli:

1. **Create the GitHub release** as normal using GoReleaser
   ```bash
   goreleaser release --clean
   ```

2. **Update the PKGBUILD** with the new version and checksums:
   ```bash
   cd aur
   ./update-pkgbuild.sh 0.14.0  # Replace with actual version
   ```

3. **Test the package locally** (if on Arch Linux):
   ```bash
   makepkg -si
   ```

4. **Update the AUR repository**:
   ```bash
   # Clone or pull the AUR repository
   git clone ssh://aur@aur.archlinux.org/axiom-bin.git /tmp/axiom-bin-aur
   # OR if already cloned:
   # cd /path/to/axiom-bin-aur && git pull

   # Copy the updated PKGBUILD
   cp PKGBUILD /tmp/axiom-bin-aur/

   # Generate .SRCINFO
   cd /tmp/axiom-bin-aur
   makepkg --printsrcinfo > .SRCINFO

   # Commit and push
   git add PKGBUILD .SRCINFO
   git commit -m "Update to version 0.14.0"
   git push
   ```

## Architecture Support

The PKGBUILD supports the following architectures:
- `x86_64` - Downloads the `linux_amd64` binary
- `aarch64` - Downloads the `linux_arm64` binary

## GoReleaser Configuration

The `.goreleaser.yaml` file is configured with:
- `skip_upload: true` - Prevents automatic upload of the generated PKGBUILD
- The generated PKGBUILD serves as a reference but isn't used directly

## Testing

To test the PKGBUILD on different architectures:

### On x86_64:
```bash
makepkg -si
```

### On aarch64:
```bash
makepkg -si
```

### Cross-architecture testing (without actual compilation):
```bash
# Test PKGBUILD generation for different architectures
makepkg --printsrcinfo CARCH=x86_64
makepkg --printsrcinfo CARCH=aarch64
```

## Troubleshooting

### Issue: Wrong architecture downloaded
- Ensure the PKGBUILD uses `source_x86_64` and `source_aarch64` arrays
- Check that the URLs correctly map architectures (amd64 → x86_64, arm64 → aarch64)

### Issue: Checksum validation fails
- Run `update-pkgbuild.sh` to update checksums
- Verify the GitHub release has the expected archive files

### Issue: Package doesn't build
- Check that all required files (LICENSE, README.md, man pages) are included in the archives
- Verify the `axiom` binary has executable permissions in the archive

## Contributing

If you need to update the PKGBUILD structure:
1. Edit the `PKGBUILD` file in this directory
2. Test thoroughly on both architectures
3. Update this README if needed
4. Submit a PR with your changes

## Future Improvements

Potential improvements to consider:
1. Contributing a fix to GoReleaser to support multi-architecture PKGBUILDs natively
2. Creating separate AUR packages per architecture (axiom-bin-x86_64, axiom-bin-aarch64)
3. Automating the AUR update process in CI/CD after releases

## References

- [Arch Linux PKGBUILD documentation](https://wiki.archlinux.org/title/PKGBUILD)
- [AUR Submission Guidelines](https://wiki.archlinux.org/title/AUR_submission_guidelines)
- [GoReleaser AUR documentation](https://goreleaser.com/customization/aur/)
