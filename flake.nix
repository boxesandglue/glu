{
  description = "glu — render Markdown, HTML or Lua straight to PDF";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule rec {
          pname = "glu";
          version = "0.0.34";
          src = self;
          vendorHash = "sha256-YDT24RKm0f9iF5bue/8RpqtI2xlNwWdMOGJx3odTKhw=";
          subPackages = [ "glu" ];
          # The golden tests compare rendered PDFs and need fonts and
          # poppler, which are not available inside the build sandbox.
          doCheck = false;
          ldflags = [
            "-s"
            "-w"
            "-X main.Version=${version}"
          ];
          meta = {
            description = "Render Markdown, HTML or Lua straight to PDF, built on the boxes and glue typesetting library";
            homepage = "https://boxesandglue.dev";
            license = pkgs.lib.licenses.mit;
            mainProgram = "glu";
          };
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [ self.packages.${system}.default ];
        };
      }
    );
}
