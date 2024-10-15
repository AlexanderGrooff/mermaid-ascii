{
  inputs = {
    alex.url = "github:AlexanderGrooff/nix-flakes";
  };

  outputs = {
    self,
    flake-utils,
    nixpkgs,
    alex,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in {
        devShells.default = alex.devShells.${system}.go;
        packages.default = pkgs.buildGoModule {
          name = "mermaid-ascii";
          src = ./.;
          vendorHash = "sha256-Lnksoj7vx3T1lJfJRuGJHqKMvwra7t7TGn8NwjEkMMI=";
        };
      }
    );
}
