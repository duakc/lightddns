{
  description = "lightddns - lightweight dynamic DNS updater";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      # Version from the flake's git revision, matching the git-based versioning
      # used everywhere else. Falls back through dirty rev to "dev".
      version = self.shortRev or self.dirtyShortRev or "dev";

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      # Files live under release/nix/; reference them through `self` (the flake
      # root) so the root flake.nix symlink doesn't skew relative paths.
      mkPackage =
        pkgs:
        pkgs.callPackage (self + "/release/nix/package.nix") {
          src = self;
          inherit version;
        };
    in
    {
      packages = forAllSystems (pkgs: rec {
        "lightddns" = mkPackage pkgs;
        default = lightddns;
      });

      apps = forAllSystems (
        pkgs:
        let
          program = nixpkgs.lib.getExe self.packages.${pkgs.system}.default;
        in
        rec {
          "lightddns" = {
            type = "app";
            inherit program;
            meta.description = "Lightweight dynamic DNS (DDNS) updater";
          };
          default = lightddns;
        }
      );

      overlays.default = _final: prev: { "lightddns" = mkPackage prev; };

      # Importable service modules; each defaults its package to this flake's
      # build for the evaluating system. NixOS uses systemd, nix-darwin launchd.
      nixosModules.default =
        { pkgs, ... }:
        {
          imports = [ (self + "/release/nix/module.nix") ];
          services."lightddns".package = nixpkgs.lib.mkDefault self.packages.${pkgs.system}.default;
        };

      darwinModules.default =
        { pkgs, ... }:
        {
          imports = [ (self + "/release/nix/darwin.nix") ];
          services."lightddns".package = nixpkgs.lib.mkDefault self.packages.${pkgs.system}.default;
        };

      formatter = forAllSystems (pkgs: pkgs.nixfmt);
    };
}
