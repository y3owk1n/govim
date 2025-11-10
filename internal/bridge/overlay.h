#ifndef OVERLAY_H
#define OVERLAY_H

#import <Foundation/Foundation.h>
#import <CoreGraphics/CoreGraphics.h>

// Overlay window handle
typedef void* OverlayWindow;

// Hint style configuration
typedef struct {
    int fontSize;
    char* fontFamily;
    char* backgroundColor;
    char* textColor;
    char* matchedTextColor;
    char* borderColor;
    int borderRadius;
    int borderWidth;
    int padding;
    double opacity;
    int showArrow;  // 0 = no arrow, 1 = show arrow
} HintStyle;

// Hint data
typedef struct {
    char* label;
    CGPoint position;
    CGSize size;
    int matchedPrefixLength;  // Number of matched characters to highlight
} HintData;

// Grid cell style configuration
typedef struct {
    int fontSize;
    char* fontFamily;
    char* backgroundColor;
    char* textColor;
    char* matchedTextColor;
    char* matchedBackgroundColor;
    char* matchedBorderColor;
    char* borderColor;
    int borderWidth;
    double backgroundOpacity;
    double textOpacity;
} GridCellStyle;

// Grid cell data
typedef struct {
    char* label;
    CGRect bounds;  // Cell rectangle
    int isMatched;  // 1 if cell matches current input, 0 otherwise
    int isSubgrid;  // 1 if cell is part of subgrid, 0 otherwise
} GridCell;

// Function declarations
OverlayWindow createOverlayWindow();
void destroyOverlayWindow(OverlayWindow window);
void showOverlayWindow(OverlayWindow window);
void hideOverlayWindow(OverlayWindow window);
void clearOverlay(OverlayWindow window);
void drawHints(OverlayWindow window, HintData* hints, int count, HintStyle style);
void drawScrollHighlight(OverlayWindow window, CGRect bounds, char* color, int width);
void setOverlayLevel(OverlayWindow window, int level);
void drawTargetDot(OverlayWindow window, CGPoint center, double radius, const char *color, const char *borderColor, double borderWidth);
void replaceOverlayWindow(OverlayWindow *pwindow);
void resizeOverlayToMainScreen(OverlayWindow window);
void resizeOverlayToActiveScreen(OverlayWindow window);

// Grid-specific drawing functions
void drawGridCells(OverlayWindow window, GridCell* cells, int count, GridCellStyle style);
void drawGridLines(OverlayWindow window, CGRect* lines, int count, char* color, int width, double opacity);
void updateGridMatchPrefix(OverlayWindow window, const char* prefix);
void setHideUnmatched(OverlayWindow window, int hide);

#endif // OVERLAY_H
