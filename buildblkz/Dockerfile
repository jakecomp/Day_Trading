FROM mcr.microsoft.com/devcontainers/go


ARG USERNAME=kube-trader && USER_UID=1000 && USER_GID=$USER_UID
#ARG USER_UID=1000
#ARG USER_GID=$USER_UID
ENV USERNAME=kube-trader

COPY . .

RUN usermod -l kube-trader vscode \
    && mv /home/vscode /home/kube-trader \
    && export DEBIAN_FRONTEND=noninteractive \
    && ./ensure-mount-sources.sh \
    && ./installdock.sh \ 
    && copy-kube-config/install.sh \
    && go install github.com/derailed/k9s@latest \
    # kubectl zsh completion
    && echo "source <(kubectl completion zsh)" >> ~/.zshrc \
    && echo "source <(helm completion zsh)" >> ~/.zshrc \
    && echo "source <(minikube completion zsh)" >> ~/.zshrc \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get autoremove -y \
    && apt-get clean -y \
    && export DEBIAN_FRONTEND=dialog


#ENV DEVCONTAINER=true
    #mkdir -p "${USERHOME}/.oh-my-zsh/completions"
    #kubectl completion zsh > "${USERHOME}/.oh-my-zsh/completions/_kubectl"
    #chown -R "${USERNAME}" "${USERHOME}/.oh-my-zsh"
    #source <(kubectl completion zsh)

# [Optional] Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends <your-package-list-here>
#RUN useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
#&& groupadd --gid $USER_GID $USERNAME 


#ARG USERNAME=kube-trader
# RUN echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
#    && chmod 0440 /etc/sudoers.d/$USERNAME
 # && curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL=latest sh - \
 #RUN curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash \
        # safe wau if above breaks
    #curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
    #chmod 700 get_helm.sh
    #./get_helm.sh
#cd .devcontainer
    # get k3's 
    #&& 