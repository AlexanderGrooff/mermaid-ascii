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
        devShells.default = pkgs.mkShell {
          nativeBuildInputs = [
            pkgs.go_1_23
            pkgs.pre-commit
            pkgs.golangci-lint

            # Add other Go tools or dependencies here if needed, e.g.:
            pkgs.gopls
            pkgs.delve
            pkgs.goreleaser
            pkgs.air
          ];

          # Disable hardening for cgo
          hardeningDisable = [ "all" ];
        };
        packages.default = pkgs.buildGoModule {
          name = "mermaid-ascii";
          src = ./.;
          vendorHash = "sha256-S/K6W8KC6YzwZPioucoiwOMd29LPv0J22T3MS0X+W5g=";
        };
      }
    );
}
