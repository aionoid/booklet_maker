{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  buildInputs = with pkgs; [
    figlet
    pdfcpu
    go
    pkg-config
    cairo
    glib
    gtk3
    libx11
    libxxf86vm
    #dev
    python3Packages.reportlab
  ];
}
