# lightddns RPM spec.
#
# Go stages the whole packaged tree (binary, config, systemd units, man page)
# into the staging dir and this spec just copies that tree into the buildroot -
# there is no prep or build step. Package metadata lives in the header fields,
# install-time behaviour in the scriptlets below.
# NOTE: keep rpm macro tokens out of these comments - rpm expands them even
# inside comments and warns.

# The binary is already stripped (-w -s) and may be a foreign arch when building
# all targets on one host, so disable rpm's debuginfo extraction and the brp
# post-install steps (host strip/compress would choke on a cross-built binary).
# We ship exactly what Go staged. The arch is set by rpmbuild's --target flag.
%global debug_package %{nil}
%global __os_install_post %{nil}

Name:           lightddns
Version:        __VERSION__
Release:        1%{?dist}
Summary:        Lightweight dynamic DNS (DDNS) updater

License:        GPL-2.0-only
URL:            https://lightddns.duaky.com

Requires(pre):    shadow-utils

%description
lightddns watches your public/local IP through pluggable datasources
and pushes changes to DNS providers such as Cloudflare.

%install
cp -a %{staging}/. %{buildroot}/

%pre
# Dedicated system group/user (no home, no login shell).
getent group lightddns >/dev/null || groupadd --system lightddns
getent passwd lightddns >/dev/null || \
    useradd --system --gid lightddns --no-create-home \
        --home-dir /var/lib/lightddns --shell /sbin/nologin \
        --comment "lightddns DDNS daemon" lightddns
exit 0

%post
# Config + secrets: readable by the service group, not world.
# (systemd also re-applies /var/lib/lightddns via StateDirectory=.)
for f in /etc/lightddns.yaml /etc/default/lightddns; do
    if [ -e "$f" ]; then
        chown lightddns:lightddns "$f"
        chmod 0600 "$f"
    fi
done

if [ -d /etc/lightddns.d ]; then
    chown lightddns:lightddns /etc/lightddns.d
    chmod 0700 /etc/lightddns.d
fi

# State/working dir for `-D` (created here so non-systemd setups work too).
if [ ! -d /var/lib/lightddns ]; then
    mkdir -p /var/lib/lightddns
fi
chown lightddns:lightddns /var/lib/lightddns
chmod 0700 /var/lib/lightddns

if [ -d /run/systemd/system ]; then
    systemctl daemon-reload || true
    systemctl enable lightddns.service || true
    # On upgrade ($1 -ge 2) restart to pick up the new binary. On a fresh
    # install we leave it stopped so the admin can fill in the config.
    if [ "$1" -ge 2 ]; then
        systemctl try-restart lightddns.service || true
    fi
fi
exit 0

%preun
# $1 == 0 is the final erase (not an upgrade).
if [ "$1" -eq 0 ] && [ -d /run/systemd/system ]; then
    systemctl --no-reload disable lightddns.service || true
    systemctl stop lightddns.service || true
fi
exit 0

%postun
if [ -d /run/systemd/system ]; then
    systemctl daemon-reload || true
    # $1 -ge 1 is an upgrade: restart into the new binary.
    if [ "$1" -ge 1 ]; then
        systemctl try-restart lightddns.service || true
    fi
fi

# Note: per Fedora packaging guidelines the system user/group is intentionally
# left behind on erase so any files it still owns keep a valid owner.
exit 0

%files
/usr/bin/lightddns
/usr/lib/systemd/system/lightddns.service
/usr/lib/systemd/system/lightddns@.service
%{_mandir}/man1/lightddns.1.gz
%dir /etc/lightddns.d
%config(noreplace) /etc/lightddns.yaml
%config(noreplace) /etc/lightddns.d/example.yaml
%config(noreplace) /etc/default/lightddns