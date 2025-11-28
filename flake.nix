{
  description = "lazyssh flake with build + development shell";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        lib = pkgs.lib;
        rev = self.rev or "dirty";
        version = "git-${rev}";
        goToolchain = if pkgs ? go_1_24 then pkgs.go_1_24 else pkgs.go;

        lazyssh = pkgs.buildGoModule {
          pname = "lazyssh";
          inherit version;
          src = ./.;
          subPackages = [ "./cmd" ];
          vendorHash = "sha256-OMlpqe7FJDqgppxt4t8lJ1KnXICOh6MXVXoKkYJ74Ks=";
          go = goToolchain;
          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
            "-X main.gitCommit=${rev}"
          ];
        };
      in {
        packages = {
          default = lazyssh;
          inherit lazyssh;
        };

        apps.default = flake-utils.lib.mkApp {
          drv = lazyssh;
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            goToolchain
            golangci-lint
            gofumpt
            gnumake
            git
          ];
        };
      });
}
