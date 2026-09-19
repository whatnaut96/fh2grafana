{
    description = "Nix shell for development";
    inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    outputs = { self, nixpkgs }:
      let
        pkgs = nixpkgs.legacyPackages.x86_64-linux;
      in {
        devShells.x86_64-linux.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.gopls
          ];
        };
      };
}
