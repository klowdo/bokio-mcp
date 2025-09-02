{
  description = "Bokio MCP Server - Model Context Protocol server for Bokio API integration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    pre-commit-hooks = {
      url = "github:cachix/pre-commit-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    { self
    , nixpkgs
    , flake-utils
    , pre-commit-hooks
    ,
    }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = nixpkgs.legacyPackages.${system};

      # Pre-commit hooks configuration
      pre-commit-check = pre-commit-hooks.lib.${system}.run {
        src = ./.;
        hooks = {
          # Use our custom make-based hooks by referencing the config file
          # The actual hooks are defined in .pre-commit-config.yaml
          nixpkgs-fmt.enable = true; # Format Nix files
          prettier = {
            enable = true;
            excludes = [ "schemas/.*\\.ya?ml" ]; # Don't format downloaded API schemas
            settings = {
              tab-width = 2;
            };
          };
        };
        # Override to use our comprehensive .pre-commit-config.yaml
        settings.ormolu.defaultExtensions = [ ];
      };

      buildInputs = with pkgs; [
        go_1_24
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
        # Pre-commit and related tools
        pre-commit
        nixpkgs-fmt
        nodePackages.prettier
      ];
    in
    {
      devShells.default = pkgs.mkShell {
        inherit buildInputs;

        shellHook = ''
          echo "🚀 Bokio MCP Server development environment"
          echo "Go version: $(go version)"
          echo ""
          echo "Available commands:"
          echo "  make help               - Show all available make targets"
          echo "  make build              - Build the MCP server"
          echo "  make test               - Run tests with coverage"
          echo "  make lint               - Run linting and formatting"
          echo "  make security           - Run security scans"
          echo ""
          echo "Pre-commit hooks:"
          echo "  make pre-commit-install - Install pre-commit hooks"
          echo "  make pre-commit-run     - Run hooks on all files"
          echo "  make pre-commit         - Run full pre-commit pipeline"
          echo ""
          ${pre-commit-check.shellHook}
        '';

        # Set up Go environment
        CGO_ENABLED = "1";
        GOROOT = "${pkgs.go_1_24}/share/go";
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
          maintainers = [ ];
        };
      };

      apps.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/bokio-mcp";
      };

      checks = {
        pre-commit-check = pre-commit-check;
      };

      formatter = pkgs.nixfmt-rfc-style;
    });
}
