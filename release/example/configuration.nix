# Copy the services."lightddns" block into your system configuration after
# importing the module from this flake:
#
#   inputs.lightddns.url = "github:duakc/lightddns";
#   # in your modules list:
#   lightddns.nixosModules.default     # NixOS
#   lightddns.darwinModules.default    # nix-darwin
#
# docs: https://lightddns.duaky.com
#
# Secrets must NOT go in `settings` (the Nix store is world-readable). Manage
# them with sops-nix and point environmentFile at the decrypted file; the
# config then reads them as {{ .Env.KEY }}:
#   services."lightddns".environmentFile = config.sops.secrets."lightddns/env".path;
{ ... }:
{
  services."lightddns" = {
    enable = true;

    settings = {
      log.level = "info";
      datasources = [ ];
      providers = [ ];
      domains = [ ];
      services = [ ];
    };
  };
}
