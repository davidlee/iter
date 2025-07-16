{
  description = "A basic flake with a shell";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  inputs.systems.url = "github:nix-systems/default";
  inputs.flake-utils = {
    url = "github:numtide/flake-utils";
    inputs.systems.follows = "systems";
  };

  outputs = {
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        devShells.default = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            # D2 diagram rendering dependencies for PNG files
            # xorg.libX11
            # xorg.libXext
            # xorg.libXrandr
            # xorg.libXcomposite
            # xorg.libXdamage
            # xorg.libXfixes
            # xorg.libxcb
            # glib
            # gobject-introspection
            # nss
            # nspr
            # dbus
            # atk
            # at-spi2-atk
            # cups
            # libdrm
            # expat
            # libxkbcommon
            # pango
            # cairo
            # systemd
            # alsa-lib
            # Playwright browsers
            vscode # for playwright
            playwright-driver.browsers
          ];

          packages = with pkgs; [
            go # compiler
            gopls # language server
            delve # debugger
            gofumpt # strict formatter
            golangci-lint # linter
            d2 # diagram tool

            structurizr-cli # for docker image > d2 diagrams (unused atmo)
          ];

          # shellHook = ''
          #   export PLAYWRIGHT_BROWSERS_PATH=${pkgs.playwright-driver.browsers}
          #   export PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=true
          #   # Set up library paths for Playwright
          #   export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath [
          #     pkgs.xorg.libX11
          #     pkgs.xorg.libXext
          #     pkgs.xorg.libXrandr
          #     pkgs.xorg.libXcomposite
          #     pkgs.xorg.libXdamage
          #     pkgs.xorg.libXfixes
          #     pkgs.xorg.libxcb
          #     pkgs.glib
          #     pkgs.gobject-introspection
          #     pkgs.nss
          #     pkgs.nspr
          #     pkgs.dbus
          #     pkgs.atk
          #     pkgs.at-spi2-atk
          #     pkgs.cups
          #     pkgs.libdrm
          #     pkgs.expat
          #     pkgs.libxkbcommon
          #     pkgs.pango
          #     pkgs.cairo
          #     pkgs.systemd
          #     pkgs.alsa-lib
          #   ]}:$LD_LIBRARY_PATH
          # '';
        };
      }
    );
}
