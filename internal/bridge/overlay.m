#import "overlay.h"
#import <Cocoa/Cocoa.h>

@interface OverlayView : NSView
@property (nonatomic, strong) NSMutableArray *hints;
@property (nonatomic, strong) NSFont *hintFont;
@property (nonatomic, strong) NSColor *hintTextColor;
@property (nonatomic, strong) NSColor *hintMatchedTextColor;
@property (nonatomic, strong) NSColor *hintBackgroundColor;
@property (nonatomic, strong) NSColor *hintBorderColor;
@property (nonatomic, assign) CGFloat hintBorderRadius;
@property (nonatomic, assign) CGFloat hintBorderWidth;
@property (nonatomic, assign) CGFloat hintPadding;
@property (nonatomic, assign) CGRect scrollHighlight;
@property (nonatomic, strong) NSColor *scrollHighlightColor;
@property (nonatomic, assign) int scrollHighlightWidth;
@property (nonatomic, assign) BOOL showScrollHighlight;
@property (nonatomic, assign) BOOL showTargetDot;
@property (nonatomic, assign) CGPoint targetDotCenter;
@property (nonatomic, assign) CGFloat targetDotRadius;
@property (nonatomic, strong) NSColor *targetDotBackgroundColor;
@property (nonatomic, strong) NSColor *targetDotBorderColor;
@property (nonatomic, assign) CGFloat targetDotBorderWidth;
- (void)applyStyle:(HintStyle)style;
- (NSColor *)colorFromHex:(NSString *)hexString defaultColor:(NSColor *)defaultColor;
@end

@implementation OverlayView

- (instancetype)initWithFrame:(NSRect)frame {
    self = [super initWithFrame:frame];
    if (self) {
        _hints = [NSMutableArray array];
        _showScrollHighlight = NO;
        _showTargetDot = NO;
        _targetDotRadius = 4.0;
        _targetDotBorderWidth = 1.0;
        _targetDotBorderColor = [NSColor blackColor];
        _targetDotBackgroundColor = [[NSColor colorWithRed:1.0 green:0.84 blue:0.0 alpha:1.0] colorWithAlphaComponent:0.95];
        _hintFont = [NSFont boldSystemFontOfSize:14.0];
        _hintTextColor = [NSColor blackColor];
        _hintMatchedTextColor = [NSColor systemBlueColor];
        _hintBackgroundColor = [[NSColor colorWithRed:1.0 green:0.84 blue:0.0 alpha:1.0] colorWithAlphaComponent:0.95];
        _hintBorderColor = [NSColor blackColor];
        _hintBorderRadius = 4.0;
        _hintBorderWidth = 1.0;
        _hintPadding = 4.0;
    }
    return self;
}

- (void)drawRect:(NSRect)dirtyRect {
    [super drawRect:dirtyRect];

    // Clear background
    [[NSColor clearColor] setFill];
    NSRectFill(dirtyRect);

    // Draw scroll highlight if enabled
    if (self.showScrollHighlight) {
        [self drawScrollHighlight];
    }

   // Draw target dot if enabled (draw before hints so hints appear on top)
    if (self.showTargetDot) {
        [self drawTargetDot];
    }

    // Draw hints
    [self drawHints];
}

- (void)applyStyle:(HintStyle)style {
    CGFloat fontSize = style.fontSize > 0 ? style.fontSize : 14.0;
    NSString *fontFamily = nil;
    if (style.fontFamily) {
        fontFamily = [NSString stringWithUTF8String:style.fontFamily];
        if (fontFamily.length == 0) {
            fontFamily = nil;
        }
    }
    NSFont *font = nil;
    if (fontFamily.length > 0) {
        font = [NSFont fontWithName:fontFamily size:fontSize];
    }
    if (!font) {
        font = [NSFont fontWithName:@"Menlo-Bold" size:fontSize];
    }
    if (!font) {
        font = [NSFont boldSystemFontOfSize:fontSize];
    }
    self.hintFont = font;

    NSColor *defaultBg = [[NSColor colorWithRed:1.0 green:0.84 blue:0.0 alpha:1.0] colorWithAlphaComponent:0.95];
    NSColor *defaultText = [NSColor blackColor];
    NSColor *defaultMatchedText = [NSColor systemBlueColor];
    NSColor *defaultBorder = [NSColor blackColor];

    NSString *backgroundHex = style.backgroundColor ? [NSString stringWithUTF8String:style.backgroundColor] : nil;
    NSString *textHex = style.textColor ? [NSString stringWithUTF8String:style.textColor] : nil;
    NSString *matchedTextHex = style.matchedTextColor ? [NSString stringWithUTF8String:style.matchedTextColor] : nil;
    NSString *borderHex = style.borderColor ? [NSString stringWithUTF8String:style.borderColor] : nil;

    NSColor *backgroundColor = [self colorFromHex:backgroundHex defaultColor:defaultBg];
    CGFloat opacity = style.opacity;
    if (opacity < 0.0 || opacity > 1.0) {
        opacity = 0.95;
    }
    self.hintBackgroundColor = [backgroundColor colorWithAlphaComponent:opacity];
    self.hintTextColor = [self colorFromHex:textHex defaultColor:defaultText];
    self.hintMatchedTextColor = [self colorFromHex:matchedTextHex defaultColor:defaultMatchedText];
    self.hintBorderColor = [self colorFromHex:borderHex defaultColor:defaultBorder];

    self.hintBorderRadius = style.borderRadius > 0 ? style.borderRadius : 4.0;
    self.hintBorderWidth = style.borderWidth > 0 ? style.borderWidth : 1.0;
    self.hintPadding = style.padding >= 0 ? style.padding : 4.0;
}

- (NSColor *)colorFromHex:(NSString *)hexString defaultColor:(NSColor *)defaultColor {
    if (!hexString || hexString.length == 0) {
        return defaultColor;
    }

    NSString *cleanString = [hexString stringByTrimmingCharactersInSet:[NSCharacterSet whitespaceAndNewlineCharacterSet]];
    if ([cleanString hasPrefix:@"#"]) {
        cleanString = [cleanString substringFromIndex:1];
    }

    if (cleanString.length != 6 && cleanString.length != 8) {
        return defaultColor;
    }

    unsigned long long hexValue = 0;
    NSScanner *scanner = [NSScanner scannerWithString:cleanString];
    if (![scanner scanHexLongLong:&hexValue]) {
        return defaultColor;
    }

    CGFloat alpha = 1.0;
    if (cleanString.length == 8) {
        alpha = ((hexValue & 0xFF000000) >> 24) / 255.0;
    }
    CGFloat red = ((hexValue & 0x00FF0000) >> 16) / 255.0;
    CGFloat green = ((hexValue & 0x0000FF00) >> 8) / 255.0;
    CGFloat blue = (hexValue & 0x000000FF) / 255.0;

    return [NSColor colorWithRed:red green:green blue:blue alpha:alpha];
}

- (NSBezierPath *)createTooltipPath:(NSRect)rect
                          arrowSize:(CGFloat)arrowSize
                     elementCenterX:(CGFloat)elementCenterX
                    elementCenterY:(CGFloat)elementCenterY {
    NSBezierPath *path = [NSBezierPath bezierPath];

    // Tooltip body rectangle (excluding arrow space)
    NSRect bodyRect = NSMakeRect(rect.origin.x, rect.origin.y, rect.size.width, rect.size.height - arrowSize);

    // Arrow dimensions
    CGFloat arrowTipX = elementCenterX;
    CGFloat arrowTipY = elementCenterY;
    CGFloat arrowBaseY = bodyRect.origin.y + bodyRect.size.height;
    CGFloat arrowWidth = arrowSize * 2.5;  // Make arrow wider for better visibility
    CGFloat arrowLeft = arrowTipX - arrowWidth / 2;
    CGFloat arrowRight = arrowTipX + arrowWidth / 2;

    // Clamp arrow to tooltip bounds
    CGFloat tooltipLeft = bodyRect.origin.x + self.hintBorderRadius;
    CGFloat tooltipRight = bodyRect.origin.x + bodyRect.size.width - self.hintBorderRadius;
    arrowLeft = MAX(arrowLeft, tooltipLeft);
    arrowRight = MIN(arrowRight, tooltipRight);
    arrowTipX = (arrowLeft + arrowRight) / 2;  // Recenter if clamped

    // Start from top-left corner
    [path moveToPoint:NSMakePoint(bodyRect.origin.x + self.hintBorderRadius, bodyRect.origin.y)];

    // Top edge
    [path lineToPoint:NSMakePoint(bodyRect.origin.x + bodyRect.size.width - self.hintBorderRadius, bodyRect.origin.y)];

    // Top-right corner
    [path appendBezierPathWithArcFromPoint:NSMakePoint(bodyRect.origin.x + bodyRect.size.width, bodyRect.origin.y)
                                   toPoint:NSMakePoint(bodyRect.origin.x + bodyRect.size.width, bodyRect.origin.y + self.hintBorderRadius)
                                    radius:self.hintBorderRadius];

    // Right edge
    [path lineToPoint:NSMakePoint(bodyRect.origin.x + bodyRect.size.width, arrowBaseY - self.hintBorderRadius)];

    // Bottom-right corner
    [path appendBezierPathWithArcFromPoint:NSMakePoint(bodyRect.origin.x + bodyRect.size.width, arrowBaseY)
                                   toPoint:NSMakePoint(bodyRect.origin.x + bodyRect.size.width - self.hintBorderRadius, arrowBaseY)
                                    radius:self.hintBorderRadius];

    // Bottom edge to arrow right side
    [path lineToPoint:NSMakePoint(arrowRight, arrowBaseY)];

    // Arrow right side to tip
    [path lineToPoint:NSMakePoint(arrowTipX, arrowTipY)];

    // Arrow left side
    [path lineToPoint:NSMakePoint(arrowLeft, arrowBaseY)];

    // Continue bottom edge to bottom-left corner
    [path lineToPoint:NSMakePoint(bodyRect.origin.x + self.hintBorderRadius, arrowBaseY)];

    // Bottom-left corner
    [path appendBezierPathWithArcFromPoint:NSMakePoint(bodyRect.origin.x, arrowBaseY)
                                   toPoint:NSMakePoint(bodyRect.origin.x, arrowBaseY - self.hintBorderRadius)
                                    radius:self.hintBorderRadius];

    // Left edge
    [path lineToPoint:NSMakePoint(bodyRect.origin.x, bodyRect.origin.y + self.hintBorderRadius)];

    // Top-left corner
    [path appendBezierPathWithArcFromPoint:NSMakePoint(bodyRect.origin.x, bodyRect.origin.y)
                                   toPoint:NSMakePoint(bodyRect.origin.x + self.hintBorderRadius, bodyRect.origin.y)
                                    radius:self.hintBorderRadius];

    [path closePath];
    return path;
}

- (void)drawHints {
    for (NSDictionary *hint in self.hints) {
        NSString *label = hint[@"label"];
        if (!label || [label length] == 0) continue;

        NSPoint position = [hint[@"position"] pointValue];
        NSNumber *matchedPrefixLengthNum = hint[@"matchedPrefixLength"];
        int matchedPrefixLength = matchedPrefixLengthNum ? [matchedPrefixLengthNum intValue] : 0;
        NSNumber *showArrowNum = hint[@"showArrow"];
        BOOL showArrow = showArrowNum ? [showArrowNum boolValue] : YES;

        // Create attributed string with matched prefix in different color
        NSMutableAttributedString *attrString = [[NSMutableAttributedString alloc] initWithString:label];
        [attrString addAttribute:NSFontAttributeName value:self.hintFont range:NSMakeRange(0, [label length])];
        [attrString addAttribute:NSForegroundColorAttributeName value:self.hintTextColor range:NSMakeRange(0, [label length])];

        // Highlight matched prefix
        if (matchedPrefixLength > 0 && matchedPrefixLength <= [label length]) {
            [attrString addAttribute:NSForegroundColorAttributeName
                               value:self.hintMatchedTextColor
                               range:NSMakeRange(0, matchedPrefixLength)];
        }

        NSSize textSize = [attrString size];

        // Calculate hint box size (include arrow space if needed)
        CGFloat padding = self.hintPadding;
        CGFloat arrowHeight = showArrow ? 2.0 : 0.0;  // Arrow height only if enabled

        // Calculate dimensions - ensure box is at least square
        CGFloat contentWidth = textSize.width + (padding * 2);
        CGFloat contentHeight = textSize.height + (padding * 2);

        // Make box square if content is narrow, otherwise use content width
        CGFloat boxWidth = MAX(contentWidth, contentHeight);
        CGFloat boxHeight = contentHeight + arrowHeight;

        // Position tooltip above element with arrow pointing down to element center
        // Element center is at (position.x + size.x/2, position.y + size.y/2)
        CGFloat elementCenterX = position.x + (hint[@"size"] ? [hint[@"size"] sizeValue].width : 0) / 2.0;
        CGFloat elementCenterY = position.y + (hint[@"size"] ? [hint[@"size"] sizeValue].height : 0) / 2.0;

        // Position tooltip body above element (arrow points down)
        CGFloat gap = 3.0;  // Small gap between arrow tip and element
        CGFloat tooltipX = elementCenterX - boxWidth / 2.0;  // Center tooltip horizontally on element
        CGFloat tooltipY = elementCenterY + arrowHeight + gap;  // Position tooltip body above element

        // Convert coordinates (macOS uses bottom-left origin, we need top-left)
        NSScreen *mainScreen = [NSScreen mainScreen];
        CGFloat screenHeight = [mainScreen frame].size.height;
        CGFloat flippedY = screenHeight - tooltipY - boxHeight;
        CGFloat flippedElementCenterY = screenHeight - elementCenterY;

        NSRect hintRect = NSMakeRect(
            tooltipX,
            flippedY,
            boxWidth,
            boxHeight
        );

        // Draw tooltip background
        NSBezierPath *path;
        if (showArrow) {
            // Draw with arrow pointing to element center
            path = [self createTooltipPath:hintRect
                                arrowSize:arrowHeight
                           elementCenterX:elementCenterX
                          elementCenterY:flippedElementCenterY];
        } else {
            // Draw simple rounded rectangle without arrow
            path = [NSBezierPath bezierPathWithRoundedRect:hintRect
                                                   xRadius:self.hintBorderRadius
                                                   yRadius:self.hintBorderRadius];
        }

        [self.hintBackgroundColor setFill];
        [path fill];

        [self.hintBorderColor setStroke];
        [path setLineWidth:self.hintBorderWidth];
        [path stroke];

        // Draw text (centered in tooltip body)
        CGFloat textX = hintRect.origin.x + (boxWidth - textSize.width) / 2.0;  // Center horizontally
        CGFloat textY = hintRect.origin.y + padding;
        NSPoint textPosition = NSMakePoint(textX, textY);
        [attrString drawAtPoint:textPosition];
    }
}

- (void)drawScrollHighlight {
    if (CGRectIsEmpty(self.scrollHighlight)) {
        return;
    }

    NSGraphicsContext *context = [NSGraphicsContext currentContext];
    [context saveGraphicsState];

    // Convert coordinates (macOS uses bottom-left origin)
    NSScreen *mainScreen = [NSScreen mainScreen];
    CGFloat screenHeight = [mainScreen frame].size.height;
    CGFloat flippedY = screenHeight - self.scrollHighlight.origin.y - self.scrollHighlight.size.height;

    NSRect rect = NSMakeRect(
        self.scrollHighlight.origin.x,
        flippedY,
        self.scrollHighlight.size.width,
        self.scrollHighlight.size.height
    );

    NSBezierPath *path = [NSBezierPath bezierPathWithRect:rect];

    if (self.scrollHighlightColor) {
        [self.scrollHighlightColor setStroke];
    } else {
        [[NSColor redColor] setStroke];
    }
    [path setLineWidth:self.scrollHighlightWidth];
    [path stroke];

    [context restoreGraphicsState];
}

- (void)drawTargetDot {
    NSGraphicsContext *context = [NSGraphicsContext currentContext];
    [context saveGraphicsState];

    // Convert coordinates (macOS uses bottom-left origin)
    NSScreen *mainScreen = [NSScreen mainScreen];
    CGFloat screenHeight = [mainScreen frame].size.height;
    CGFloat flippedY = screenHeight - self.targetDotCenter.y;

    CGFloat x = self.targetDotCenter.x - self.targetDotRadius;
    CGFloat y = flippedY - self.targetDotRadius;
    CGFloat diameter = self.targetDotRadius * 2;

    NSRect dotRect = NSMakeRect(x, y, diameter, diameter);
    NSBezierPath *circlePath = [NSBezierPath bezierPathWithOvalInRect:dotRect];

    // Fill the dot
    if (self.targetDotBackgroundColor) {
        [self.targetDotBackgroundColor setFill];
    } else {
        [[NSColor redColor] setFill];
    }
    [circlePath fill];

    if (self.targetDotBorderColor && self.targetDotBorderWidth > 0) {
        [self.targetDotBorderColor setStroke];
        [circlePath setLineWidth:self.targetDotBorderWidth];
        [circlePath stroke];
    }

    [context restoreGraphicsState];
}

- (NSColor *)colorFromHex:(NSString *)hexString {
    if (!hexString || [hexString length] == 0) {
        return [NSColor blackColor];
    }

    unsigned rgbValue = 0;
    NSScanner *scanner = [NSScanner scannerWithString:hexString];
    if ([hexString hasPrefix:@"#"]) {
        [scanner setScanLocation:1]; // Skip '#'
    }
    [scanner scanHexInt:&rgbValue];

    return [NSColor colorWithRed:((rgbValue & 0xFF0000) >> 16)/255.0
                           green:((rgbValue & 0xFF00) >> 8)/255.0
                            blue:(rgbValue & 0xFF)/255.0
                           alpha:1.0];
}

@end

@interface OverlayWindowController : NSObject
@property (nonatomic, strong) NSWindow *window;
@property (nonatomic, strong) OverlayView *overlayView;
@end

@implementation OverlayWindowController

- (instancetype)init {
    self = [super init];
    if (self) {
        [self createWindow];
    }
    return self;
}

- (void)createWindow {
    NSScreen *mainScreen = [NSScreen mainScreen];
    NSRect screenFrame = [mainScreen frame];

    // Create window
    self.window = [[NSWindow alloc] initWithContentRect:screenFrame
                                              styleMask:NSWindowStyleMaskBorderless
                                                backing:NSBackingStoreBuffered
                                                  defer:NO];

    if ([self.window respondsToSelector:@selector(setAnimationBehavior:)]) {
        [self.window setAnimationBehavior:NSWindowAnimationBehaviorNone];
    }
    [self.window setAnimations:@{}];
    [self.window setAlphaValue:1.0];

    [self.window setLevel:NSScreenSaverWindowLevel]; // Higher level to be above everything
    [self.window setOpaque:NO];
    [self.window setBackgroundColor:[NSColor clearColor]];
    [self.window setIgnoresMouseEvents:YES];
    [self.window setAcceptsMouseMovedEvents:NO];
    [self.window setHasShadow:NO];
    [self.window setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces |
                                        NSWindowCollectionBehaviorStationary |
                                        NSWindowCollectionBehaviorFullScreenAuxiliary |
                                        NSWindowCollectionBehaviorIgnoresCycle];

    // Create overlay view
    self.overlayView = [[OverlayView alloc] initWithFrame:screenFrame];
    [self.window setContentView:self.overlayView];
}

@end

// C interface implementation

OverlayWindow createOverlayWindow() {
    OverlayWindowController *controller = [[OverlayWindowController alloc] init];
    return (void*)controller;
}

void destroyOverlayWindow(OverlayWindow window) {
    if (!window) return;

    OverlayWindowController *controller = (OverlayWindowController*)window;
    [controller.window close];
    [controller release];
}

void showOverlayWindow(OverlayWindow window) {
    if (!window) return;

    OverlayWindowController *controller = (OverlayWindowController*)window;

    dispatch_async(dispatch_get_main_queue(), ^{
        // Set window level to maximum
        [controller.window setLevel:kCGMaximumWindowLevel];

        // Set collection behavior
        [controller.window setCollectionBehavior:
            NSWindowCollectionBehaviorCanJoinAllSpaces |
            NSWindowCollectionBehaviorStationary |
            NSWindowCollectionBehaviorIgnoresCycle |
            NSWindowCollectionBehaviorFullScreenAuxiliary];

        // orderFrontRegardless to make it visible
        [controller.window setIsVisible:YES];
        [controller.window orderFrontRegardless];
        [controller.window makeKeyAndOrderFront:nil];

        // Force display update
        [controller.window display];
        [controller.overlayView setNeedsDisplay:YES];
    });
}

void hideOverlayWindow(OverlayWindow window) {
    if (!window) return;

    OverlayWindowController *controller = (OverlayWindowController*)window;

    if ([NSThread isMainThread]) {
        [controller.window orderOut:nil];
    } else {
        dispatch_async(dispatch_get_main_queue(), ^{
            [controller.window orderOut:nil];
        });
    }
}

void clearOverlay(OverlayWindow window) {
    if (!window) return;

    OverlayWindowController *controller = (OverlayWindowController*)window;

    if ([NSThread isMainThread]) {
        [controller.overlayView.hints removeAllObjects];
        controller.overlayView.showScrollHighlight = NO;
        controller.overlayView.showTargetDot = NO;
        [controller.overlayView setNeedsDisplay:YES];
    } else {
        dispatch_async(dispatch_get_main_queue(), ^{
            [controller.overlayView.hints removeAllObjects];
            controller.overlayView.showScrollHighlight = NO;
            controller.overlayView.showTargetDot = NO;
            [controller.overlayView setNeedsDisplay:YES];
        });
    }
}

void drawHints(OverlayWindow window, HintData* hints, int count, HintStyle style) {
    if (!window || !hints) return;

    OverlayWindowController *controller = (OverlayWindowController*)window;

    if ([NSThread isMainThread]) {
        // Execute directly - original code
        [controller.overlayView.hints removeAllObjects];
        [controller.overlayView applyStyle:style];

        for (int i = 0; i < count; i++) {
            HintData hint = hints[i];
            NSDictionary *hintDict = @{
                @"label": @(hint.label),
                @"position": [NSValue valueWithPoint:NSPointFromCGPoint(hint.position)],
                @"matchedPrefixLength": @(hint.matchedPrefixLength),
                @"showArrow": @(style.showArrow)
            };
            [controller.overlayView.hints addObject:hintDict];
        }

        [controller.overlayView setNeedsDisplay:YES];
    } else {
        // Need to copy data for async dispatch
        NSMutableArray *hintDicts = [NSMutableArray arrayWithCapacity:count];
        for (int i = 0; i < count; i++) {
            HintData hint = hints[i];
            NSDictionary *hintDict = @{
                @"label": @(hint.label),
                @"position": [NSValue valueWithPoint:NSPointFromCGPoint(hint.position)],
                @"matchedPrefixLength": @(hint.matchedPrefixLength),
                @"showArrow": @(style.showArrow)
            };
            [hintDicts addObject:hintDict];
        }

        // Deep copy style strings
        HintStyle styleCopy;
        styleCopy.fontSize = style.fontSize;
        styleCopy.borderRadius = style.borderRadius;
        styleCopy.borderWidth = style.borderWidth;
        styleCopy.padding = style.padding;
        styleCopy.opacity = style.opacity;
        styleCopy.showArrow = style.showArrow;
        styleCopy.fontFamily = style.fontFamily ? strdup(style.fontFamily) : NULL;
        styleCopy.backgroundColor = style.backgroundColor ? strdup(style.backgroundColor) : NULL;
        styleCopy.textColor = style.textColor ? strdup(style.textColor) : NULL;
        styleCopy.matchedTextColor = style.matchedTextColor ? strdup(style.matchedTextColor) : NULL;
        styleCopy.borderColor = style.borderColor ? strdup(style.borderColor) : NULL;

        dispatch_async(dispatch_get_main_queue(), ^{
            [controller.overlayView.hints removeAllObjects];
            [controller.overlayView applyStyle:styleCopy];
            [controller.overlayView.hints addObjectsFromArray:hintDicts];
            [controller.overlayView setNeedsDisplay:YES];

            // Free copied strings
            if (styleCopy.fontFamily) free((void*)styleCopy.fontFamily);
            if (styleCopy.backgroundColor) free((void*)styleCopy.backgroundColor);
            if (styleCopy.textColor) free((void*)styleCopy.textColor);
            if (styleCopy.matchedTextColor) free((void*)styleCopy.matchedTextColor);
            if (styleCopy.borderColor) free((void*)styleCopy.borderColor);
        });
    }
}

void drawScrollHighlight(OverlayWindow window, CGRect bounds, char* color, int width) {
    if (!window) return;

    OverlayWindowController *controller = (OverlayWindowController*)window;

    if ([NSThread isMainThread]) {
        // Execute directly - original code
        controller.overlayView.scrollHighlight = bounds;
        controller.overlayView.scrollHighlightWidth = width;
        controller.overlayView.showScrollHighlight = YES;

        if (color) {
            NSString *colorStr = @(color);
            unsigned rgbValue = 0;
            NSScanner *scanner = [NSScanner scannerWithString:colorStr];
            [scanner setScanLocation:1];
            [scanner scanHexInt:&rgbValue];

            controller.overlayView.scrollHighlightColor = [NSColor colorWithRed:((rgbValue & 0xFF0000) >> 16)/255.0
                                                                          green:((rgbValue & 0xFF00) >> 8)/255.0
                                                                           blue:(rgbValue & 0xFF)/255.0
                                                                          alpha:1.0];
        }

        [controller.overlayView setNeedsDisplay:YES];
    } else {
        NSString *colorStr = color ? [NSString stringWithUTF8String:color] : nil;

        dispatch_async(dispatch_get_main_queue(), ^{
            controller.overlayView.scrollHighlight = bounds;
            controller.overlayView.scrollHighlightWidth = width;
            controller.overlayView.showScrollHighlight = YES;

            if (colorStr) {
                unsigned rgbValue = 0;
                NSScanner *scanner = [NSScanner scannerWithString:colorStr];
                [scanner setScanLocation:1];
                [scanner scanHexInt:&rgbValue];

                controller.overlayView.scrollHighlightColor = [NSColor colorWithRed:((rgbValue & 0xFF0000) >> 16)/255.0
                                                                              green:((rgbValue & 0xFF00) >> 8)/255.0
                                                                               blue:(rgbValue & 0xFF)/255.0
                                                                              alpha:1.0];
            }

            [controller.overlayView setNeedsDisplay:YES];
        });
    }
}

void drawTargetDot(OverlayWindow window, CGPoint center, double radius, const char *colorStr, const char *borderColorStr, double borderWidth) {
    if (!window) return;

    OverlayWindowController *controller = (OverlayWindowController*)window;

    if ([NSThread isMainThread]) {
        // Execute directly - original code
        controller.overlayView.targetDotCenter = center;
        controller.overlayView.targetDotRadius = radius;
        controller.overlayView.targetDotBorderWidth = borderWidth;
        controller.overlayView.showTargetDot = YES;

        if (colorStr) {
            NSString *colorString = @(colorStr);
            controller.overlayView.targetDotBackgroundColor = [controller.overlayView colorFromHex:colorString
                                                                                    defaultColor:[NSColor redColor]];
        } else {
            controller.overlayView.targetDotBackgroundColor = [NSColor redColor];
        }

        if (borderColorStr) {
            NSString *borderColorString = @(borderColorStr);
            controller.overlayView.targetDotBorderColor = [controller.overlayView colorFromHex:borderColorString
                                                                                 defaultColor:[NSColor blackColor]];
        } else {
            controller.overlayView.targetDotBorderColor = nil;
        }

        [controller.overlayView setNeedsDisplay:YES];
    } else {
        NSString *colorString = colorStr ? [NSString stringWithUTF8String:colorStr] : nil;
        NSString *borderColorString = borderColorStr ? [NSString stringWithUTF8String:borderColorStr] : nil;

        dispatch_async(dispatch_get_main_queue(), ^{
            controller.overlayView.targetDotCenter = center;
            controller.overlayView.targetDotRadius = radius;
            controller.overlayView.targetDotBorderWidth = borderWidth;
            controller.overlayView.showTargetDot = YES;

            if (colorString) {
                controller.overlayView.targetDotBackgroundColor = [controller.overlayView colorFromHex:colorString
                                                                                        defaultColor:[NSColor redColor]];
            } else {
                controller.overlayView.targetDotBackgroundColor = [NSColor redColor];
            }

            if (borderColorString) {
                controller.overlayView.targetDotBorderColor = [controller.overlayView colorFromHex:borderColorString
                                                                                     defaultColor:[NSColor blackColor]];
            } else {
                controller.overlayView.targetDotBorderColor = nil;
            }

            [controller.overlayView setNeedsDisplay:YES];
        });
    }
}

void setOverlayLevel(OverlayWindow window, int level) {
    if (!window) return;

    OverlayWindowController *controller = (OverlayWindowController*)window;

    if ([NSThread isMainThread]) {
        [controller.window setLevel:level];
    } else {
        dispatch_async(dispatch_get_main_queue(), ^{
            [controller.window setLevel:level];
        });
    }
}
