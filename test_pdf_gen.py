# This script requires the ReportLab library.
# Install it using: pip install reportlab

from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import A4
from reportlab.lib.units import inch
from reportlab.lib.utils import simpleSplit
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont


def generate_number_test_pdf(filename="book.pdf", num_pages=200):
    """
    Generates a PDF document where each page contains the current number
    printed once horizontally (Portrait) and once vertically (Landscape)
    for comprehensive printing program testing.
    Also includes an up arrow (▲) centered on each page.

    Args:
        filename (str): The name of the output PDF file.
        num_pages (int): The total number of pages/numbers to generate.
    """
    print(f"Starting PDF generation for {num_pages} pages...")

    # 0. Register Font
    # pdfmetrics.registerFont(
    #     TTFont("Noto Color Emoji Regular", "./NotoColorEmoji-Regular.ttf")
    # )
    FONT_NAME = "BigBlueTermPlus Nerd Font Mono Regular"
    pdfmetrics.registerFont(
        TTFont(
            FONT_NAME,
            "./fonts/BigBlueTermPlusNerdFontMono-Regular.ttf",
        )
    )

    # 1. Initialize the Canvas
    c = canvas.Canvas(filename, pagesize=A4)
    w, h = A4  # Get page width and height

    # 2. Define drawing constants
    # FONT_NAME = "Helvetica-Bold"
    # FONT_NAME = "Noto Color Emoji Regular"

    # Font size for the large numbers (200pt)
    NUM_FONT_SIZE = 150
    # Font size for the center arrow (100pt)
    ARROW_FONT_SIZE = 250

    # 3. Loop through the required number range
    for i in range(1, num_pages + 1):
        num_str = str(i)

        # --- Draw 1 & 2: Horizontal and Vertical Numbers ---
        c.setFont(FONT_NAME, NUM_FONT_SIZE)

        # Draw 1: Horizontal (Portrait Orientation)
        # Place near the top center of the page
        c.drawCentredString(w / 2, h * 0.7, num_str)

        # Draw 2: Vertical (Landscape Orientation)
        c.saveState()  # Save the current canvas state before rotation/translation

        # Target the left-center of the page (e.g., 25% across width, 50% up height)
        c.translate(w * 0.25, h / 2)
        c.rotate(90)
        c.drawCentredString(0, 0, num_str)

        c.restoreState()  # Restore the canvas state (undoes translate and rotate)

        # --- Draw 3: Center Arrow (Up Arrow) ---
        c.setFont(FONT_NAME, ARROW_FONT_SIZE)

        # Draw the triangle (▲) at the absolute center (w/2, h/2)
        # Using drawCentredString centers it both horizontally and vertically (visually)
        c.drawCentredString(w / 2, h / 2.5, "󰜸")

        # End the current page and start a new one for the next number
        c.showPage()

    # 4. Save the document
    try:
        c.save()
        print(f"\nSuccessfully generated '{filename}' with {num_pages} pages.")
        print(
            "Each page now contains the number horizontally, vertically, and a center arrow."
        )
    except Exception as e:
        print(f"An error occurred during PDF creation: {e}")


if __name__ == "__main__":
    # Generate the PDF with numbers 1 through 200
    generate_number_test_pdf(num_pages=200)
