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

      # Schema files are now stored locally in the repository
      # Use `make update-schema` to download the latest versions

      # Pre-commit hooks configuration using git-hooks.nix
      # Using minimal set of known working hooks
      pre-commit-check = pre-commit-hooks.lib.${system}.run {
        src = ./.;
        hooks = {
          # Nix formatting
          nixpkgs-fmt.enable = true;

          # Go formatting
          gofmt.enable = true;

          # YAML/JSON formatting (excluding downloaded schemas)
          prettier = {
            enable = true;
            excludes = [ "schemas/.*\\.ya?ml" ];
            settings = {
              tab-width = 2;
            };
          };
        };
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
        # Code formatting tools (used by git-hooks)
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
          echo "Code Quality:"
          echo "  nix flake check         - Run git-hooks and quality checks"
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

        vendorHash = "sha256-5TdupBsoknikVrc4qShgDzZEuCaifHyG4PcC+WO7ng8=";

        # Build with Go 1.24
        nativeBuildInputs = [ pkgs.go_1_24 ];

        # Set build flags
        ldflags = [
          "-s"
          "-w"
          "-X main.version=${self.packages.${system}.default.version or "nix-build"}"
        ];

        # Ensure we use the correct Go version
        preBuild = ''
          export GOROOT="${pkgs.go_1_24}/share/go"
        '';

        meta = with pkgs.lib; {
          description = "Bokio MCP Server - Model Context Protocol server for Bokio API integration";
          homepage = "https://github.com/klowdo/bokio-mcp";
          license = licenses.mit;
          maintainers = [ ];
          platforms = platforms.unix;
        };
      };

      apps.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/bokio-mcp";
      };

      checks = {
        inherit pre-commit-check;
      };

      formatter = pkgs.nixfmt-rfc-style;
    });
}
