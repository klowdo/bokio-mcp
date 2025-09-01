{
  description = "Bokio MCP Server - Model Context Protocol server for Bokio API integration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};

      buildInputs = with pkgs; [
        go_1_23
        gotools
        gopls
        delve
        golangci-lint
        govulncheck
        gosec
        # gomod2nix
        goreleaser
        oapi-codegen
        git
        gnumake
        curl
        jq
      ];
    in {
      devShells.default = pkgs.mkShell {
        inherit buildInputs;

        shellHook = ''
          echo "🚀 Bokio MCP Server development environment"
          echo "Go version: $(go version)"
          echo ""
          echo "Available commands:"
          echo "  make help     - Show all available make targets"
          echo "  go mod init   - Initialize Go module"
          echo "  make build    - Build the MCP server"
          echo "  make test     - Run tests with coverage"
          echo "  make lint     - Run linting and formatting"
          echo ""
        '';

        # Set up Go environment
        CGO_ENABLED = "1";
        GOROOT = "${pkgs.go_1_23}/share/go";
        GOPROXY = "https://proxy.golang.org,direct";
        GOSUMDB = "sum.golang.org";
      };

      packages.default = pkgs.buildGoModule {
        pname = "bokio-mcp";
        version = "0.1.0";

        src = ./.;

        # This will be updated after go.mod is created
        vendorHash = null;

        meta = with pkgs.lib; {
          description = "Model Context Protocol server for Bokio API integration";
          homepage = "https://github.com/klowdo/bokio-mcp";
          license = licenses.mit;
          maintainers = [];
        };
      };

      apps.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/bokio-mcp";
      };

      formatter = pkgs.nixfmt-rfc-style;
    });
}
