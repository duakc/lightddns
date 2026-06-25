# Shared option declarations for the lightddns service. Both the NixOS (systemd)
# and nix-darwin (launchd) modules import this, so the option surface is
# identical on either platform - you flip `services."lightddns".enable` the same
# way regardless of OS.
{ lib, pkgs, ... }:
{
  options.services."lightddns" = {
    enable = lib.mkEnableOption "lightddns dynamic DNS updater";

    package = lib.mkPackageOption pkgs "lightddns" { };

    settings = lib.mkOption {
      type = (pkgs.formats.yaml { }).type;
      default = {
        log.level = "info";
      };
      example = lib.literalExpression ''
        {
          log.level = "info";
          # providers / datasources / domains / services go here;
          # see the docs for the full schema.
        }
      '';
      description = ''
        Daemon configuration, rendered to a YAML file and passed to lightddns.
        This is the primary way to configure the service on Nix.

        Minimal starting template:
        ```nix
        services."lightddns".settings = {
          log.level = "info";
        };
        ```

        For the full configuration schema and examples, see the documentation:
        <https://lightddns.duaky.com>.

        Do NOT put secrets here - the Nix store is world-readable. Reference them
        as `{{ .Env.KEY }}` and supply them via `environmentFile`, or manage
        them with sops-nix (see `environmentFile`). Ignored when `configFile` is
        set.
      '';
    };

    configFile = lib.mkOption {
      type = lib.types.nullOr lib.types.path;
      default = null;
      description = ''
        Use this config file verbatim instead of rendering `settings`. Handy
        when the file is generated for you - e.g. a sops-nix template that inlines
        secrets directly into the YAML.
      '';
    };

    environmentFile = lib.mkOption {
      type = lib.types.nullOr lib.types.path;
      default = null;
      example = "/run/secrets/lightddns.env";
      description = ''
        Path to an environment file with the secrets the config references as
        `{{ .Env.KEY }}`. Keep it out of the Nix store and manage it with
        sops-nix, e.g.:
        ```nix
        sops.secrets."lightddns/env" = { };
        services."lightddns".environmentFile =
          config.sops.secrets."lightddns/env".path;
        ```
      '';
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "lightddns";
      description = "User the service runs as.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "lightddns";
      description = "Group the service runs as.";
    };

    extraArgs = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [ ];
      example = [ "--once" ];
      description = "Extra arguments appended to the `run` sub-command.";
    };
  };
}
