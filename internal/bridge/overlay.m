#import "overlay.h"
#import <Cocoa/Cocoa.h>

@interface OverlayView : NSView
@property (nonatomic, strong) NSMutableArray *hints;
@property (nonatomic, assign) HintStyle style;
@property (nonatomic, assign) CGRect scrollHighlight;
@property (nonatomic, strong) NSColor *scrollHighlightColor;
@property (nonatomic, assign) int scrollHighlightWidth;
@property (nonatomic, assign) BOOL showScrollHighlight;
@end

@implementation OverlayView

- (instancetype)initWithFrame:(NSRect)frame {
    self = [super initWithFrame:frame];
    if (self) {
        _hints = [NSMutableArray array];
        _showScrollHighlight = NO;
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

- (void)drawHints {
    for (NSDictionary *hint in self.hints) {
        NSString *label = hint[@"label"];
        if (!label || [label length] == 0) continue;
        
        NSPoint position = [hint[@"position"] pointValue];
        
        // Use system font
        NSFont *font = [NSFont boldSystemFontOfSize:14];
        
        // Simple colors - yellow background, black text
        NSColor *textColor = [NSColor blackColor];
        NSColor *bgColor = [NSColor colorWithRed:1.0 green:0.84 blue:0.0 alpha:0.95]; // Gold
        
        NSDictionary *attributes = @{
            NSFontAttributeName: font,
            NSForegroundColorAttributeName: textColor
        };
        
        NSAttributedString *attrString = [[NSAttributedString alloc] initWithString:label attributes:attributes];
        NSSize textSize = [attrString size];
        
        // Calculate hint box size
        CGFloat padding = 4;
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
        NSBezierPath *path = [NSBezierPath bezierPathWithRoundedRect:hintRect xRadius:4 yRadius:4];
        [bgColor setFill];
        [path fill];
        
        [[NSColor blackColor] setStroke];
        [path setLineWidth:1];
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
    [controller.window orderFront:nil];
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
    
    // Set style
    controller.overlayView.style = style;
    
    // Add new hints
    for (int i = 0; i < count; i++) {
        HintData hint = hints[i];
        
        NSDictionary *hintDict = @{
            @"label": @(hint.label),
            @"position": [NSValue valueWithPoint:NSPointFromCGPoint(hint.position)]
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
