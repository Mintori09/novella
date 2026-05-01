{
  description = "novella";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      repoRoot = ./.;
      frontendRoot = repoRoot + "/frontend";
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
      mkLinuxDeps =
        pkgs:
        with pkgs;
        lib.optionals stdenv.isLinux [
          gtk3
          webkitgtk_4_1
          glib
          gdk-pixbuf
          pango
          cairo
          nss
          atk
          at-spi2-atk
          libxkbcommon
          libx11
          libxcursor
          libxrandr
          libxi
          libxext
          libxdamage
          libxfixes
          libxcomposite
          libxrender
          libxtst
          libxcb
        ];
      mkFrontendProd =
        pkgs:
        pkgs.buildNpmPackage {
          pname = "novella-frontend";
          version = "0.1.0";
          src = frontendRoot;

          npmDeps = pkgs.importNpmLock {
            npmRoot = frontendRoot;
          };

          npmConfigHook = pkgs.importNpmLock.npmConfigHook;
          npmBuildScript = "build";

          installPhase = ''
            runHook preInstall
            mkdir -p $out/dist
            cp -R dist/. $out/dist/
            runHook postInstall
          '';
        };
      mkNovellaProd =
        pkgs:
        pkgs.stdenv.mkDerivation {
          pname = "novella";
          version = "0.1.0";
          src = repoRoot;

          nativeBuildInputs = with pkgs; [
            go
            nodejs_22
            wails
            pkg-config
            gcc
          ];

          buildInputs = mkLinuxDeps pkgs;

          env = {
            CGO_ENABLED = "1";
            HOME = "/build";
          };

          buildPhase = ''
            runHook preBuild
            rm -rf frontend/dist
            mkdir -p frontend/dist
            cp -R ${mkFrontendProd pkgs}/dist/. frontend/dist/
            substituteInPlace wails.json \
              --replace '"frontend:install": "npm install"' '"frontend:install": "true"' \
              --replace '"frontend:build": "npm run build"' '"frontend:build": "true"'
            wails build -clean
            runHook postBuild
          '';

          installPhase = ''
            runHook preInstall
            install -Dm755 build/bin/novella $out/bin/novella
            runHook postInstall
          '';
        };
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = mkNovellaProd pkgs;
        }
      );

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/novella";
        };
      });

      devShells = forAllSystems (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
          novellaProd = mkNovellaProd pkgs;
        in
        pkgs.mkShell {
          packages =
            with pkgs;
            [
              go
              nodejs_22
              wails
              pkg-config
              gcc
            ]
            ++ mkLinuxDeps pkgs;

          shellHook = ''
            export CGO_ENABLED=1
            if [ -z "''${NOVELLA_NO_AUTORUN:-}" ]; then
              exec ${novellaProd}/bin/novella
            fi
          '';
        }
      );
    };
}
