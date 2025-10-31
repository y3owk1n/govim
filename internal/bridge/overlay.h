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

// Function declarations
OverlayWindow createOverlayWindow();
void destroyOverlayWindow(OverlayWindow window);
void showOverlayWindow(OverlayWindow window);
void hideOverlayWindow(OverlayWindow window);
void clearOverlay(OverlayWindow window);
void drawHints(OverlayWindow window, HintData* hints, int count, HintStyle style);
void drawScrollHighlight(OverlayWindow window, CGRect bounds, char* color, int width);
void setOverlayLevel(OverlayWindow window, int level);

#endif // OVERLAY_H
