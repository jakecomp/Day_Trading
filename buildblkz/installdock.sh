#!/usr/bin/env bash

DOCKER_VERSION="${VERSION:-"latest"}"
USE_MOBY="${MOBY:-"true"}"
DOCKER_DASH_COMPOSE_VERSION="${DOCKERDASHCOMPOSEVERSION:-"v2"}" # v1 or v2

ENABLE_NONROOT_DOCKER="${ENABLE_NONROOT_DOCKER:-"true"}"
SOURCE_SOCKET="${SOURCE_SOCKET:-"/var/run/docker-host.sock"}"
TARGET_SOCKET="${TARGET_SOCKET:-"/var/run/docker.sock"}"
USERNAME="${USERNAME:-"${_REMOTE_USER:-"automatic"}"}"
INSTALL_DOCKER_BUILDX="${INSTALLDOCKERBUILDX:-"true"}"

MICROSOFT_GPG_KEYS_URI="https://packages.microsoft.com/keys/microsoft.asc"
DOCKER_MOBY_ARCHIVE_VERSION_CODENAMES="buster bullseye bionic focal jammy"
DOCKER_LICENSED_ARCHIVE_VERSION_CODENAMES="buster bullseye bionic focal hirsute impish jammy"

set -e

# Clean up
rm -rf /var/lib/apt/lists/*

if [ "$(id -u)" -ne 0 ]; then
    echo -e 'Script must be run as root. Use sudo, su, or add "USER root" to your Dockerfile before running this script.'
    exit 1
fi

# Determine the appropriate non-root user
if [ "${USERNAME}" = "auto" ] || [ "${USERNAME}" = "automatic" ]; then
    USERNAME=""
    POSSIBLE_USERS=("kube-trader" "node" "codespace" "$(awk -v val=1000 -F ":" '$3==val{print $1}' /etc/passwd)")
    for CURRENT_USER in "${POSSIBLE_USERS[@]}"; do
        if id -u ${CURRENT_USER} > /dev/null 2>&1; then
            USERNAME=${CURRENT_USER}
            break
        fi
    done
    if [ "${USERNAME}" = "" ]; then
        USERNAME=kube-trader
    fi
elif [ "${USERNAME}" = "none" ] || ! id -u ${USERNAME} > /dev/null 2>&1; then
    USERNAME=kube-trader
fi
USERHOME="/home/$USERNAME"
KUBECTL_VERSION="${VERSION:-"latest"}"
HELM_VERSION="${HELM:-"latest"}"
MINIKUBE_VERSION="${MINIKUBE:-"latest"}"

KUBECTL_SHA256="${KUBECTL_SHA256:-"automatic"}"
HELM_SHA256="${HELM_SHA256:-"automatic"}"
MINIKUBE_SHA256="${MINIKUBE_SHA256:-"automatic"}"
USERNAME="${USERNAME:-"${_REMOTE_USER:-"automatic"}"}"

HELM_GPG_KEYS_URI="https://raw.githubusercontent.com/helm/helm/main/KEYS"
# GPG_KEY_SERVERS="keyserver hkp://keyserver.ubuntu.com"
# keyserver hkps://keys.openpgp.org
# keyserver hkp://keyserver.pgp.com


# Get central common setting
get_common_setting() {
    if [ "${common_settings_file_loaded}" != "true" ]; then
        curl -sfL "https://aka.ms/vscode-dev-containers/script-library/settings.env" 2>/dev/null -o /tmp/vsdc-settings.env || echo "Could not download settings file. Skipping."
        common_settings_file_loaded=true
    fi
    if [ -f "/tmp/vsdc-settings.env" ]; then
        local multi_line=""
        if [ "$2" = "true" ]; then multi_line="-z"; fi
        local result="$(grep ${multi_line} -oP "$1=\"?\K[^\"]+" /tmp/vsdc-settings.env | tr -d '\0')"
        if [ ! -z "${result}" ]; then declare -g $1="${result}"; fi
    fi
    echo "$1=${!1}"
}

apt_get_update()
{
    if [ "$(find /var/lib/apt/lists/* | wc -l)" = "0" ]; then
        echo "Running apt-get update..."
        apt-get update -y
    fi
}

# Checks if packages are installed and installs them if not
check_packages() {
    if ! dpkg -s "$@" > /dev/null 2>&1; then
        apt_get_update
        apt-get -y install --no-install-recommends "$@"
    fi
}

# Figure out correct version of a three part version number is not passed
find_version_from_git_tags() {
    local variable_name=$1
    local requested_version=${!variable_name}
    if [ "${requested_version}" = "none" ]; then return; fi
    local repository=$2
    local prefix=${3:-"tags/v"}
    local separator=${4:-"."}
    local last_part_optional=${5:-"false"}    
    if [ "$(echo "${requested_version}" | grep -o "." | wc -l)" != "2" ]; then
        local escaped_separator=${separator//./\\.}
        local last_part
        if [ "${last_part_optional}" = "true" ]; then
            last_part="(${escaped_separator}[0-9]+)?"
        else
            last_part="${escaped_separator}[0-9]+"
        fi
        local regex="${prefix}\\K[0-9]+${escaped_separator}[0-9]+${last_part}$"
        local version_list="$(git ls-remote --tags ${repository} | grep -oP "${regex}" | tr -d ' ' | tr "${separator}" "." | sort -rV)"
        if [ "${requested_version}" = "latest" ] || [ "${requested_version}" = "current" ] || [ "${requested_version}" = "lts" ]; then
            declare -g ${variable_name}="$(echo "${version_list}" | head -n 1)"
        else
            set +e
            declare -g ${variable_name}="$(echo "${version_list}" | grep -E -m 1 "^${requested_version//./\\.}([\\.\\s]|$)")"
            set -e
        fi
    fi
    if [ -z "${!variable_name}" ] || ! echo "${version_list}" | grep "^${!variable_name//./\\.}$" > /dev/null 2>&1; then
        echo -e "Invalid ${variable_name} value: ${requested_version}\nValid values:\n${version_list}" >&2
        exit 1
    fi
    echo "${variable_name}=${!variable_name}"
}

# Ensure apt is in non-interactive to avoid prompts
export DEBIAN_FRONTEND=noninteractive

# Install dependencies
check_packages apt-transport-https curl ca-certificates gnupg2 dirmngr wget curl ca-certificates coreutils gnupg2 dirmngr bash-completion
if ! type git > /dev/null 2>&1; then
    check_packages git
fi

# Source /etc/os-release to get OS info
. /etc/os-release
# Fetch host/container arch.
architecture="$(dpkg --print-architecture)"

if [ ${KUBECTL_VERSION} != "none" ]; then
    # Install the kubectl, verify checksum
    echo "Downloading kubectl..."
    if [ "${KUBECTL_VERSION}" = "latest" ] || [ "${KUBECTL_VERSION}" = "lts" ] || [ "${KUBECTL_VERSION}" = "current" ] || [ "${KUBECTL_VERSION}" = "stable" ]; then
        KUBECTL_VERSION="$(curl -sSL https://dl.k8s.io/release/stable.txt)"
    else
        find_version_from_git_tags KUBECTL_VERSION https://github.com/kubernetes/kubernetes
    fi
    if [ "${KUBECTL_VERSION::1}" != 'v' ]; then
        KUBECTL_VERSION="v${KUBECTL_VERSION}"
    fi
    curl -sSL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${architecture}/kubectl"
    chmod 0755 /usr/local/bin/kubectl
    if [ "$KUBECTL_SHA256" = "automatic" ]; then
        KUBECTL_SHA256="$(curl -sSL "https://dl.k8s.io/${KUBECTL_VERSION}/bin/linux/${architecture}/kubectl.sha256")"
    fi
    ([ "${KUBECTL_SHA256}" = "dev-mode" ] || (echo "${KUBECTL_SHA256} */usr/local/bin/kubectl" | sha256sum -c -))
    if ! type kubectl > /dev/null 2>&1; then
        echo '(!) kubectl installation failed!'
        exit 1
    fi

    # kubectl bash completion
    kubectl completion bash > /etc/bash_completion.d/kubectl

    # kubectl zsh completion
    if [ -e "${USERHOME}}/.oh-my-zsh" ]; then
        mkdir -p "${USERHOME}/.oh-my-zsh/completions"
        kubectl completion zsh > "${USERHOME}/.oh-my-zsh/completions/_kubectl"
        chown -R "${USERNAME}" "${USERHOME}/.oh-my-zsh"
        #source <(kubectl completion zsh)

    fi
fi


# Check if distro is supported
if [ "${USE_MOBY}" = "true" ]; then
    # 'get_common_setting' allows attribute to be updated remotely
    get_common_setting DOCKER_MOBY_ARCHIVE_VERSION_CODENAMES
    if [[ "${DOCKER_MOBY_ARCHIVE_VERSION_CODENAMES}" != *"${VERSION_CODENAME}"* ]]; then
        err "Unsupported  distribution version '${VERSION_CODENAME}'. To resolve, either: (1) set feature option '\"moby\": false' , or (2) choose a compatible OS distribution"
        exit 1
    fi
    echo "Distro codename  '${VERSION_CODENAME}'  matched filter  '${DOCKER_MOBY_ARCHIVE_VERSION_CODENAMES}'"
else
    get_common_setting DOCKER_LICENSED_ARCHIVE_VERSION_CODENAMES
    if [[ "${DOCKER_LICENSED_ARCHIVE_VERSION_CODENAMES}" != *"${VERSION_CODENAME}"* ]]; then
        err "Unsupported distribution version '${VERSION_CODENAME}'. To resolve, please choose a compatible OS distribution"
        err "Support distributions include:  ${DOCKER_LICENSED_ARCHIVE_VERSION_CODENAMES}"
        exit 1
    fi
    echo "Distro codename  '${VERSION_CODENAME}'  matched filter  '${DOCKER_LICENSED_ARCHIVE_VERSION_CODENAMES}'"
fi

# Set up the necessary apt repos (either Microsoft's or Docker's)
if [ "${USE_MOBY}" = "true" ]; then

    cli_package_name="moby-cli"

    # Import key safely and import Microsoft apt repo
    get_common_setting MICROSOFT_GPG_KEYS_URI
    curl -sSL ${MICROSOFT_GPG_KEYS_URI} | gpg --dearmor > /usr/share/keyrings/microsoft-archive-keyring.gpg
    echo "deb [arch=${architecture} signed-by=/usr/share/keyrings/microsoft-archive-keyring.gpg] https://packages.microsoft.com/repos/microsoft-${ID}-${VERSION_CODENAME}-prod ${VERSION_CODENAME} main" > /etc/apt/sources.list.d/microsoft.list
else
    # Name of proprietary engine package
    cli_package_name="docker-ce-cli"

    # Import key safely and import Docker apt repo
    curl -fsSL https://download.docker.com/linux/${ID}/gpg | gpg --dearmor > /usr/share/keyrings/docker-archive-keyring.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/${ID} ${VERSION_CODENAME} stable" > /etc/apt/sources.list.d/docker.list
fi

# Refresh apt lists
apt-get update

# Soft version matching for CLI
if [ "${DOCKER_VERSION}" = "latest" ] || [ "${DOCKER_VERSION}" = "lts" ] || [ "${DOCKER_VERSION}" = "stable" ]; then
    # Empty, meaning grab whatever "latest" is in apt repo
    cli_version_suffix=""
else    
    # Fetch a valid version from the apt-cache (eg: the Microsoft repo appends +azure, breakfix, etc...)
    docker_version_dot_escaped="${DOCKER_VERSION//./\\.}"
    docker_version_dot_plus_escaped="${docker_version_dot_escaped//+/\\+}"
    # Regex needs to handle debian package version number format: https://www.systutorials.com/docs/linux/man/5-deb-version/
    docker_version_regex="^(.+:)?${docker_version_dot_plus_escaped}([\\.\\+ ~:-]|$)"
    set +e # Don't exit if finding version fails - will handle gracefully
    cli_version_suffix="=$(apt-cache madison ${cli_package_name} | awk -F"|" '{print $2}' | sed -e 's/^[ \t]*//' | grep -E -m 1 "${docker_version_regex}")"
    set -e
    if [ -z "${cli_version_suffix}" ] || [ "${cli_version_suffix}" = "=" ]; then
        echo "(!) No full or partial Docker / Moby version match found for \"${DOCKER_VERSION}\" on OS ${ID} ${VERSION_CODENAME} (${architecture}). Available versions:"
        apt-cache madison ${cli_package_name} | awk -F"|" '{print $2}' | grep -oP '^(.+:)?\K.+'
        exit 1
    fi
    echo "cli_version_suffix ${cli_version_suffix}"
fi

# Install Docker / Moby CLI if not already installed
if type docker > /dev/null 2>&1; then
    echo "Docker / Moby CLI already installed."
else
    if [ "${USE_MOBY}" = "true" ]; then
        apt-get -y install --no-install-recommends moby-cli${cli_version_suffix} moby-buildx
        apt-get -y install --no-install-recommends moby-compose || echo "(*) Package moby-compose (Docker Compose v2) not available for OS ${ID} ${VERSION_CODENAME} (${architecture}). Skipping."
    else
        apt-get -y install --no-install-recommends docker-ce-cli${cli_version_suffix}
    fi
fi

# Install Docker Compose if not already installed  and is on a supported architecture
if type docker-compose > /dev/null 2>&1; then
    echo "Docker Compose already installed."
else
    TARGET_COMPOSE_ARCH="$(uname -m)"
    if [ "${TARGET_COMPOSE_ARCH}" = "amd64" ]; then
        TARGET_COMPOSE_ARCH="x86_64"
    fi
    if [ "${TARGET_COMPOSE_ARCH}" != "x86_64" ]; then
        # Use pip to get a version that runs on this architecture
        check_packages python3-minimal python3-pip libffi-dev python3-venv
        export PIPX_HOME=/usr/local/pipx
        mkdir -p ${PIPX_HOME}
        export PIPX_BIN_DIR=/usr/local/bin
        export PYTHONUSERBASE=/tmp/pip-tmp
        export PIP_CACHE_DIR=/tmp/pip-tmp/cache
        pipx_bin=pipx
        if ! type pipx > /dev/null 2>&1; then
            pip3 install --disable-pip-version-check --no-cache-dir --user pipx
            pipx_bin=/tmp/pip-tmp/bin/pipx
        fi
        ${pipx_bin} install --pip-args '--no-cache-dir --force-reinstall' docker-compose
        rm -rf /tmp/pip-tmp
    else 
        compose_v1_version="1"
        find_version_from_git_tags compose_v1_version "https://github.com/docker/compose" "tags/"
        echo "(*) Installing docker-compose ${compose_v1_version}..."
        curl -fsSL "https://github.com/docker/compose/releases/download/${compose_v1_version}/docker-compose-Linux-x86_64" -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
    fi
fi

# Install docker-compose switch if not already installed - https://github.com/docker/compose-switch#manual-installation
current_v1_compose_path="$(which docker-compose)"
target_v1_compose_path="$(dirname "${current_v1_compose_path}")/docker-compose-v1"
if ! type compose-switch > /dev/null 2>&1; then
    echo "(*) Installing compose-switch..."
    compose_switch_version="latest"
    find_version_from_git_tags compose_switch_version "https://github.com/docker/compose-switch"
    curl -fsSL "https://github.com/docker/compose-switch/releases/download/v${compose_switch_version}/docker-compose-linux-${architecture}" -o /usr/local/bin/compose-switch
    chmod +x /usr/local/bin/compose-switch
    # TODO: Verify checksum once available: https://github.com/docker/compose-switch/issues/11

    # Setup v1 CLI as alternative in addition to compose-switch (which maps to v2)
    mv "${current_v1_compose_path}" "${target_v1_compose_path}"
    update-alternatives --install /usr/local/bin/docker-compose docker-compose /usr/local/bin/compose-switch 99
    update-alternatives --install /usr/local/bin/docker-compose docker-compose "${target_v1_compose_path}" 1
fi
if [ "${DOCKER_DASH_COMPOSE_VERSION}" = "v1" ]; then
    update-alternatives --set docker-compose "${target_v1_compose_path}"
else
    update-alternatives --set docker-compose /usr/local/bin/compose-switch
fi

# Setup a docker group in the event the docker socket's group is not root
if ! grep -qE '^docker:' /etc/group; then
    groupadd --system docker
fi
usermod -aG docker "${USERNAME}"

if [ "${INSTALL_DOCKER_BUILDX}" = "true" ]; then
    buildx_version="latest"
    find_version_from_git_tags buildx_version "https://github.com/docker/buildx" "refs/tags/v"

    echo "(*) Installing buildx ${buildx_version}..."
    buildx_file_name="buildx-v${buildx_version}.linux-${architecture}"
    cd /tmp && wget "https://github.com/docker/buildx/releases/download/v${buildx_version}/${buildx_file_name}"

    mkdir -p ${_REMOTE_USER_HOME}/.docker/cli-plugins
    mv ${buildx_file_name} ${_REMOTE_USER_HOME}/.docker/cli-plugins/docker-buildx
    chmod +x ${_REMOTE_USER_HOME}/.docker/cli-plugins/docker-buildx

    chown -R "${USERNAME}:docker" "${_REMOTE_USER_HOME}/.docker"
    chmod -R g+r+w "${_REMOTE_USER_HOME}/.docker"
    find "${_REMOTE_USER_HOME}/.docker" -type d -print0 | xargs -n 1 -0 chmod g+s
fi

# If init file already exists, exit
if [ -f "/usr/local/share/docker-init.sh" ]; then
    # Clean up
    rm -rf /var/lib/apt/lists/*
    exit 0
fi
echo "docker-init doesn't exist, adding..."

# By default, make the source and target sockets the same
if [ "${SOURCE_SOCKET}" != "${TARGET_SOCKET}" ]; then
    touch "${SOURCE_SOCKET}"
    ln -s "${SOURCE_SOCKET}" "${TARGET_SOCKET}"
fi

if [ ${HELM_VERSION} != "none" ]; then
    # Install Helm, verify signature and checksum
    echo "Downloading Helm..."
    find_version_from_git_tags HELM_VERSION "https://github.com/helm/helm"
    if [ "${HELM_VERSION::1}" != 'v' ]; then
        HELM_VERSION="v${HELM_VERSION}"
    fi
    mkdir -p /tmp/helm
    helm_filename="helm-${HELM_VERSION}-linux-${architecture}.tar.gz"
    tmp_helm_filename="/tmp/helm/${helm_filename}"
    curl -sSL "https://get.helm.sh/${helm_filename}" -o "${tmp_helm_filename}"
    curl -sSL "https://github.com/helm/helm/releases/download/${HELM_VERSION}/${helm_filename}.asc" -o "${tmp_helm_filename}.asc"
    export GNUPGHOME="/tmp/helm/gnupg"
    mkdir -p "${GNUPGHOME}"
    chmod 700 ${GNUPGHOME}
    curl -sSL "${HELM_GPG_KEYS_URI}" -o /tmp/helm/KEYS
    echo -e "disable-ipv6\n${GPG_KEY_SERVERS}" > ${GNUPGHOME}/dirmngr.conf
    gpg -q --import "/tmp/helm/KEYS"
    if ! gpg --verify "${tmp_helm_filename}.asc" > ${GNUPGHOME}/verify.log 2>&1; then
        echo "Verification failed!"
        cat /tmp/helm/gnupg/verify.log
        exit 1
    fi

    if [ "${HELM_SHA256}" = "automatic" ]; then
        curl -sSL "https://get.helm.sh/${helm_filename}.sha256" -o "${tmp_helm_filename}.sha256"
        curl -sSL "https://github.com/helm/helm/releases/download/${HELM_VERSION}/${helm_filename}.sha256.asc" -o "${tmp_helm_filename}.sha256.asc"
        if ! gpg --verify "${tmp_helm_filename}.sha256.asc" > /tmp/helm/gnupg/verify.log 2>&1; then
            echo "Verification failed!"
            cat /tmp/helm/gnupg/verify.log
            exit 1
        fi
        HELM_SHA256="$(cat "${tmp_helm_filename}.sha256")"
    fi

    ([ "${HELM_SHA256}" = "dev-mode" ] || (echo "${HELM_SHA256} *${tmp_helm_filename}" | sha256sum -c -))
    tar xf "${tmp_helm_filename}" -C /tmp/helm
    mv -f "/tmp/helm/linux-${architecture}/helm" /usr/local/bin/
    chmod 0755 /usr/local/bin/helm
    rm -rf /tmp/helm
    if ! type helm > /dev/null 2>&1; then
        echo '(!) Helm installation failed!'
        exit 1
    fi
fi

#Install Minikube, verify checksum
if [ "${MINIKUBE_VERSION}" != "none" ]; then
    echo "Downloading minikube..."
    if [ "${MINIKUBE_VERSION}" = "latest" ] || [ "${MINIKUBE_VERSION}" = "lts" ] || [ "${MINIKUBE_VERSION}" = "current" ] || [ "${MINIKUBE_VERSION}" = "stable" ]; then
        MINIKUBE_VERSION="latest"
    else
        find_version_from_git_tags MINIKUBE_VERSION https://github.com/kubernetes/minikube
        if [ "${MINIKUBE_VERSION::1}" != "v" ]; then
            MINIKUBE_VERSION="v${MINIKUBE_VERSION}"
        fi
    fi
    # latest is also valid in the download URLs 
    curl -sSL -o /usr/local/bin/minikube "https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-${architecture}"    
    chmod 0755 /usr/local/bin/minikube
    if [ "$MINIKUBE_SHA256" = "automatic" ]; then
        MINIKUBE_SHA256="$(curl -sSL "https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-${architecture}.sha256")"
    fi
    ([ "${MINIKUBE_SHA256}" = "dev-mode" ] || (echo "${MINIKUBE_SHA256} */usr/local/bin/minikube" | sha256sum -c -))
    if ! type minikube > /dev/null 2>&1; then
        echo '(!) minikube installation failed!'
        exit 1
    fi
    # Create minkube folder with correct privs in case a volume is mounted here
    mkdir -p "${USERHOME}/.minikube"
    chown -R $USERNAME "${USERHOME}/.minikube"
    chmod -R u+wrx "${USERHOME}/.minikube"
fi

if ! type docker > /dev/null 2>&1; then
    echo -e '\n(*) Warning: The docker command was not found.\n\nYou can use one of the following scripts to install it:\n\nhttps://github.com/microsoft/vscode-dev-containers/blob/main/script-library/docs/docker-in-docker.md\n\nor\n\nhttps://github.com/microsoft/vscode-dev-containers/blob/main/script-library/docs/docker.md'
fi

# Add a stub if not adding non-root user access, user is root
if [ "${ENABLE_NONROOT_DOCKER}" = "false" ] || [ "${USERNAME}" = "root" ]; then
    echo -e '#!/usr/bin/env bash\nexec "$@"' > /usr/local/share/docker-init.sh
    chmod +x /usr/local/share/docker-init.sh
    # Clean up
    rm -rf /var/lib/apt/lists/*
    exit 0
fi

DOCKER_GID="$(grep -oP '^docker:x:\K[^:]+' /etc/group)"

# If enabling non-root access and specified user is found, setup socat and add script
chown -h "${USERNAME}":root "${TARGET_SOCKET}"        
check_packages socat
tee /usr/local/share/docker-init.sh > /dev/null \
<< EOF 
#!/usr/bin/env bash
#-------------------------------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See https://go.microsoft.com/fwlink/?linkid=2090316 for license information.
#-------------------------------------------------------------------------------------------------------------

set -e

SOCAT_PATH_BASE=/tmp/vscr-docker-from-docker
SOCAT_LOG=\${SOCAT_PATH_BASE}.log
SOCAT_PID=\${SOCAT_PATH_BASE}.pid

# Wrapper function to only use sudo if not already root
sudoIf()
{
    if [ "\$(id -u)" -ne 0 ]; then
        sudo "\$@"
    else
        "\$@"
    fi
}

# Log messages
log()
{
    echo -e "[\$(date)] \$@" | sudoIf tee -a \${SOCAT_LOG} > /dev/null
}

echo -e "\n** \$(date) **" | sudoIf tee -a \${SOCAT_LOG} > /dev/null
log "Ensuring ${USERNAME} has access to ${SOURCE_SOCKET} via ${TARGET_SOCKET}"

# If enabled, try to update the docker group with the right GID. If the group is root, 
# fall back on using socat to forward the docker socket to another unix socket so 
# that we can set permissions on it without affecting the host.
if [ "${ENABLE_NONROOT_DOCKER}" = "true" ] && [ "${SOURCE_SOCKET}" != "${TARGET_SOCKET}" ] && [ "${USERNAME}" != "root" ] && [ "${USERNAME}" != "0" ]; then
    SOCKET_GID=\$(stat -c '%g' ${SOURCE_SOCKET})
    if [ "\${SOCKET_GID}" != "0" ] && [ "\${SOCKET_GID}" != "${DOCKER_GID}" ] && ! grep -E ".+:x:\${SOCKET_GID}" /etc/group; then
        sudoIf groupmod --gid "\${SOCKET_GID}" docker
    else
        # Enable proxy if not already running
        if [ ! -f "\${SOCAT_PID}" ] || ! ps -p \$(cat \${SOCAT_PID}) > /dev/null; then
            log "Enabling socket proxy."
            log "Proxying ${SOURCE_SOCKET} to ${TARGET_SOCKET} for vscode"
            sudoIf rm -rf ${TARGET_SOCKET}
            (sudoIf socat UNIX-LISTEN:${TARGET_SOCKET},fork,mode=660,user=${USERNAME} UNIX-CONNECT:${SOURCE_SOCKET} 2>&1 | sudoIf tee -a \${SOCAT_LOG} > /dev/null & echo "\$!" | sudoIf tee \${SOCAT_PID} > /dev/null)
        else
            log "Socket proxy already running."
        fi
    fi
    log "Success"
fi

# Execute whatever commands were passed in (if any). This allows us 
# to set this script to ENTRYPOINT while still executing the default CMD.
set +e
exec "\$@"
EOF
chmod +x /usr/local/share/docker-init.sh
chown ${USERNAME}:root /usr/local/share/docker-init.sh

# Clean up
rm -rf /var/lib/apt/lists/*

echo "Done!"