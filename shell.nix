{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  buildInputs = with pkgs; [
    figlet
    pdfcpu
    #dev
    # python3Packages.reportlab
  ];
}
