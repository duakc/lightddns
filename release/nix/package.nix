# lightddns package derivation. buildGoModule compiles the daemon in the nix
# sandbox; postInstall lays down the shared release assets (systemd units, man
# page, example config) the same way the other package formats ship them.
{
  lib,
  buildGoModule,
  installShellFiles,
  # Repo root, passed from the flake as `self`, so the shared release/ assets
  # are read from one place regardless of where this file lives.
  src,
  version ? "dev",
  # Provenance stamp. A flake references a revision, not a branch, so default to
  # "nix" rather than inventing a branch name.
  branch ? "nix",
}:
buildGoModule {
  pname = "lightddns";
  inherit version src;

  # Fixed-output hash of the module dependencies; changes only when go.sum does.
  # Regenerate by setting this to lib.fakeHash and pasting the hash nix reports.
  vendorHash = "sha256-zkZ5QkBVSLKG5Bkdcgru4is7rEgHbpuilarBW2iau+k=";

  # Only the daemon. The script/goscript codegen tree is a separate module.
  subPackages = [ "cmd/lightddns" ];

  env.CGO_ENABLED = 0;

  ldflags = [
    "-s"
    "-w"
    "-X github.com/duakc/lightddns/constant.Version=${version}"
    "-X github.com/duakc/lightddns/constant.Branch=${branch}"
  ];

  nativeBuildInputs = [ installShellFiles ];

  # Ship the systemd units (the NixOS module reuses them via systemd.packages)
  # and the man page. No example config: on Nix the service is configured
  # declaratively through `services."lightddns".settings`.
  postInstall = ''
    install -Dm644 ${src}/release/systemd/lightddns.service \
      $out/lib/systemd/system/lightddns.service
    install -Dm644 ${src}/release/systemd/lightddns@.service \
      $out/lib/systemd/system/lightddns@.service

    installManPage ${src}/release/man/lightddns.1
  '';

  meta = {
    description = "Lightweight dynamic DNS (DDNS) updater";
    homepage = "https://lightddns.duaky.com";
    license = lib.licenses.gpl2Only;
    mainProgram = "lightddns";
    platforms = lib.platforms.unix;
  };
}
