{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  packages = with pkgs; [
    go # compiler
    gopls # language server
    delve # debugger
    gofumpt # strict formatter
  ];
}
