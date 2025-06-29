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

      forAllSystems = nixpkgs.lib.genAttrs nativeSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });

    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
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
