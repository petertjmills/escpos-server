{
  description = "A simple Go package for escpos with libusb";

  inputs.nixpkgs.url = "nixpkgs/nixos-25.05";

  outputs =
    { self, nixpkgs }:
    let
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
      version = builtins.substring 0 8 lastModifiedDate;

      nativeSystems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
        "aarch64-linux"
      ];

      crossSystem = {
        config = "aarch64-unknown-linux-gnu";
        libc = "glibc";
      };

      forAllSystems = nixpkgs.lib.genAttrs nativeSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });

      pkgsCross = import nixpkgs {
        system = "aarch64-linux"; # adjust if you’re building on aarch64 host
        crossSystem = crossSystem;
      };

    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          escpos-cross = pkgsCross.buildGoModule {
            pname = "escpos";
            inherit version;
            src = ./.;
            nativeBuildInputs = with pkgsCross; [ pkg-config ];
            buildInputs = with pkgsCross; [ libusb1 ];
            vendorHash = "sha256-YS+N+jTFGQpMQKczTKrZ741vhuSgszrENadpE5tbLOE=";
            CGO_ENABLED = "1";
            goFlags = [ "-v" ];
          };
          escpos = pkgs.buildGoModule {
            pname = "escpos";
            inherit version;
            src = ./.;
            nativeBuildInputs = with pkgs; [ pkg-config ];
            buildInputs = with pkgs; [ libusb1 ];
            vendorHash = "sha256-YS+N+jTFGQpMQKczTKrZ741vhuSgszrENadpE5tbLOE=";
            CGO_ENABLED = "1";
            goFlags = [ "-v" ];
          };
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = [
              pkgs.pkg-config
              pkgs.gcc
              pkgs.libusb1
              pkgs.go
              pkgs.gopls
              pkgs.gotools
              pkgs.go-tools
            ];
          };
        }
      );

      defaultPackage = forAllSystems (system: self.packages.${system}.escpos);
    };
}
