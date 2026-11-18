{
  description = "go develop";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems = nixpkgs.lib.genAttrs systems;

      # The locked nixpkgs only provides Go 1.27rc1; use the stable release.
      goFor = pkgs: pkgs.go_1_27.overrideAttrs {
        version = "1.27.3";
        src = pkgs.fetchurl {
          url = "https://go.dev/dl/go1.27.3.src.tar.gz";
          hash = "sha256-bndJA1zzYPBrJXXRYup535/ajvZxiJ5cN1+OUHd/Jrs=";
        };
      };
    in
    {
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              (goFor pkgs)
              gopls
            ];
          };
        }
      );

      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = (pkgs.buildGoModule.override { go = goFor pkgs; }) {
            pname = "NeRugaysyaBot";
            version = "0.0.1";
            src = ./.;
            vendorHash = "sha256-fws/s4VtJ9Uexk+djA1CGhEjljfCdYbGICLRTSZSiUo=";
            buildInputs = [ ];
          };
        }
      );
    };
}
