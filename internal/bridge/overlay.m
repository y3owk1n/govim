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
- (void)applyStyle:(HintStyle)style;
- (NSColor *)colorFromHex:(NSString *)hexString defaultColor:(NSColor *)defaultColor;
@end

@implementation OverlayView

- (instancetype)initWithFrame:(NSRect)frame {
    self = [super initWithFrame:frame];
    if (self) {
        _hints = [NSMutableArray array];
        _showScrollHighlight = NO;
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
    NSColor *defaultBorder = [NSColor blackColor];

    NSString *backgroundHex = style.backgroundColor ? [NSString stringWithUTF8String:style.backgroundColor] : nil;
    NSString *textHex = style.textColor ? [NSString stringWithUTF8String:style.textColor] : nil;
    NSString *borderHex = style.borderColor ? [NSString stringWithUTF8String:style.borderColor] : nil;

    NSColor *backgroundColor = [self colorFromHex:backgroundHex defaultColor:defaultBg];
    CGFloat opacity = style.opacity;
    if (opacity < 0.0 || opacity > 1.0) {
        opacity = 0.95;
    }
    self.hintBackgroundColor = [backgroundColor colorWithAlphaComponent:opacity];
    self.hintTextColor = [self colorFromHex:textHex defaultColor:defaultText];
    self.hintBorderColor = [self colorFromHex:borderHex defaultColor:defaultBorder];

    // Derive matched prefix color by blending toward system blue
    self.hintMatchedTextColor = [self.hintTextColor blendedColorWithFraction:0.4 ofColor:[NSColor systemBlueColor]];
    if (!self.hintMatchedTextColor) {
        self.hintMatchedTextColor = [NSColor systemBlueColor];
    }

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

- (void)drawHints {
    for (NSDictionary *hint in self.hints) {
        NSString *label = hint[@"label"];
        if (!label || [label length] == 0) continue;
        
        NSPoint position = [hint[@"position"] pointValue];
        NSNumber *matchedPrefixLengthNum = hint[@"matchedPrefixLength"];
        int matchedPrefixLength = matchedPrefixLengthNum ? [matchedPrefixLengthNum intValue] : 0;
        
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
        
        // Calculate hint box size
        CGFloat padding = self.hintPadding;
        CGFloat boxWidth = textSize.width + (padding * 2);
        CGFloat boxHeight = textSize.height + (padding * 2);
        
        // Convert coordinates (macOS uses bottom-left origin, we need top-left)
        NSScreen *mainScreen = [NSScreen mainScreen];
        CGFloat screenHeight = [mainScreen frame].size.height;
        CGFloat flippedY = screenHeight - position.y - boxHeight;
        
        NSRect hintRect = NSMakeRect(
            position.x,
            flippedY,
            boxWidth,
            boxHeight
        );
        
        // Draw background with border
        NSBezierPath *path = [NSBezierPath bezierPathWithRoundedRect:hintRect xRadius:self.hintBorderRadius yRadius:self.hintBorderRadius];
        [self.hintBackgroundColor setFill];
        [path fill];
        
        [self.hintBorderColor setStroke];
        [path setLineWidth:self.hintBorderWidth];
        [path stroke];
        
        // Draw text (coordinates are already flipped for the box)
        NSPoint textPosition = NSMakePoint(
            position.x + padding,
            flippedY + padding
        );
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
    [controller.window orderFrontRegardless];
}

void hideOverlayWindow(OverlayWindow window) {
    if (!window) return;
    
    OverlayWindowController *controller = (OverlayWindowController*)window;
    [controller.window orderOut:nil];
}

void clearOverlay(OverlayWindow window) {
    if (!window) return;
    
    OverlayWindowController *controller = (OverlayWindowController*)window;
    [controller.overlayView.hints removeAllObjects];
    controller.overlayView.showScrollHighlight = NO;
    [controller.overlayView setNeedsDisplay:YES];
}

void drawHints(OverlayWindow window, HintData* hints, int count, HintStyle style) {
    if (!window || !hints) return;
    
    OverlayWindowController *controller = (OverlayWindowController*)window;
    
    // Clear existing hints
    [controller.overlayView.hints removeAllObjects];

    // Apply style
    [controller.overlayView applyStyle:style];
    
    // Add new hints
    for (int i = 0; i < count; i++) {
        HintData hint = hints[i];
        
        NSDictionary *hintDict = @{
            @"label": @(hint.label),
            @"position": [NSValue valueWithPoint:NSPointFromCGPoint(hint.position)],
            @"matchedPrefixLength": @(hint.matchedPrefixLength)
        };
        
        [controller.overlayView.hints addObject:hintDict];
    }
    
    // Redraw
    [controller.overlayView setNeedsDisplay:YES];
}

void drawScrollHighlight(OverlayWindow window, CGRect bounds, char* color, int width) {
    if (!window) return;
    
    OverlayWindowController *controller = (OverlayWindowController*)window;
    
    controller.overlayView.scrollHighlight = bounds;
    controller.overlayView.scrollHighlightWidth = width;
    controller.overlayView.showScrollHighlight = YES;
    
    if (color) {
        NSString *colorStr = @(color);
        unsigned rgbValue = 0;
        NSScanner *scanner = [NSScanner scannerWithString:colorStr];
        [scanner setScanLocation:1]; // Skip '#'
        [scanner scanHexInt:&rgbValue];
        
        controller.overlayView.scrollHighlightColor = [NSColor colorWithRed:((rgbValue & 0xFF0000) >> 16)/255.0
                                                                      green:((rgbValue & 0xFF00) >> 8)/255.0
                                                                       blue:(rgbValue & 0xFF)/255.0
                                                                      alpha:1.0];
    }
    
    [controller.overlayView setNeedsDisplay:YES];
}

void setOverlayLevel(OverlayWindow window, int level) {
    if (!window) return;
    
    OverlayWindowController *controller = (OverlayWindowController*)window;
    [controller.window setLevel:level];
}
