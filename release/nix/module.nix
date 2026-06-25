# NixOS module: runs lightddns as a hardened systemd service.
#
# The hardening is NOT re-declared here. The unit shipped in the package
# (release/systemd/lightddns.service) is pulled in via systemd.packages, and a
# drop-in overrides only the parts that are FHS-specific in that file: the
# binary path (the store path, not /usr/bin), the rendered config, and the
# secrets file. So the unit stays the single source of truth for hardening.
#
# Enable with:
#   services."lightddns".enable = true;
#   services."lightddns".settings = { log.level = "info"; };       # see docs
#   services."lightddns".environmentFile = config.sops.secrets."lightddns/env".path;
{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services."lightddns";
  yaml = pkgs.formats.yaml { };
  configPath =
    if cfg.configFile != null then cfg.configFile else yaml.generate "lightddns.yaml" cfg.settings;
  exe = lib.getExe cfg.package;
in
{
  imports = [ ./options.nix ];

  config = lib.mkIf cfg.enable {
    users.users = lib.mkIf (cfg.user == "lightddns") {
      "lightddns" = {
        isSystemUser = true;
        group = cfg.group;
        home = "/var/lib/lightddns";
        description = "lightddns DDNS daemon";
      };
    };
    users.groups = lib.mkIf (cfg.group == "lightddns") { "lightddns" = { }; };

    # Bring in the hardened unit from the package ...
    systemd.packages = [ cfg.package ];

    # ... and override only what must differ on NixOS, as a drop-in on top of it.
    # A leading "" resets the unit's value before ours (systemd drop-in rule).
    systemd.services."lightddns" = {
      overrideStrategy = "asDropin";
      wantedBy = [ "multi-user.target" ];
      serviceConfig = {
        User = cfg.user;
        Group = cfg.group;
        ExecStartPre = [
          ""
          "${exe} -D /var/lib/lightddns check -c ${configPath}"
        ];
        ExecStart = [
          ""
          "${exe} -D /var/lib/lightddns run -c ${configPath} ${lib.escapeShellArgs cfg.extraArgs}"
        ];
        # Drop the unit's /etc/default reference; add the sops-managed file if set.
        EnvironmentFile = [ "" ] ++ lib.optional (cfg.environmentFile != null) cfg.environmentFile;
      };
    };
  };
}
