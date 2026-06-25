# nix-darwin module: runs lightddns as a launchd daemon on macOS.
#
# macOS has no systemd, so the systemd hardening does not apply here; this is
# the launchd equivalent of the NixOS module, driven by the same option
# (`services."lightddns".enable`).
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
  # macOS-native locations (Apple's, not Linux FHS): daemon state under
  # /Library/Application Support, logs under /Library/Logs.
  stateDir = "/Library/Application Support/lightddns";
  logFile = "/Library/Logs/lightddns.log";
in
{
  imports = [ ./options.nix ];

  config = lib.mkIf cfg.enable {
    # launchd has no StateDirectory=, so create the working dir at activation.
    system.activationScripts.extraActivation.text = lib.mkAfter ''
      mkdir -p ${lib.escapeShellArg stateDir}
    '';

    launchd.daemons."lightddns".serviceConfig = {
      ProgramArguments = [
        exe
        "-D"
        stateDir
      ]
      ++ lib.optionals (cfg.environmentFile != null) [
        "--env-file"
        (toString cfg.environmentFile)
      ]
      ++ [
        "run"
        "-c"
        (toString configPath)
      ]
      ++ cfg.extraArgs;
      RunAtLoad = true;
      KeepAlive = true;
      StandardOutPath = logFile;
      StandardErrorPath = logFile;
    };
  };
}
