#import <Cocoa/Cocoa.h>
#include <pthread.h>
#include <stdio.h>

@interface GrokBuildSwitchAppDelegate : NSObject <NSApplicationDelegate>
@property(nonatomic, copy) NSString *webURL;
- (instancetype)initWithWebURL:(NSString *)webURL;
- (void)openWebInterface:(id)sender;
- (void)requestExit;
@end

@implementation GrokBuildSwitchAppDelegate

- (instancetype)initWithWebURL:(NSString *)webURL {
    self = [super init];
    if (self) {
        _webURL = [webURL copy];
    }
    return self;
}

- (void)openWebInterface:(id)sender {
    NSURL *url = [NSURL URLWithString:self.webURL];
    if (url != nil) {
        [[NSWorkspace sharedWorkspace] openURL:url];
    }
}

- (void)requestExit {
    [NSApp stop:nil];
    NSEvent *wakeEvent = [NSEvent otherEventWithType:NSEventTypeApplicationDefined
                                            location:NSZeroPoint
                                       modifierFlags:0
                                           timestamp:0
                                        windowNumber:0
                                             context:nil
                                             subtype:0
                                               data1:0
                                               data2:0];
    [NSApp postEvent:wakeEvent atStart:NO];
}

- (BOOL)applicationShouldHandleReopen:(NSApplication *)sender hasVisibleWindows:(BOOL)hasVisibleWindows {
    [self openWebInterface:nil];
    return YES;
}

- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)sender {
    [self requestExit];
    return NSTerminateCancel;
}

@end

static GrokBuildSwitchAppDelegate *grokBuildSwitchDelegate;

static NSString *macapp_application_name(void) {
    NSString *name = [[NSBundle mainBundle] objectForInfoDictionaryKey:@"CFBundleDisplayName"];
    return name.length > 0 ? name : @"Grok Build Switch";
}

int macapp_prepare(const char *web_url) {
    if (pthread_main_np() == 0) {
        return -1;
    }
    @autoreleasepool {
        @try {
            NSApplication *app = [NSApplication sharedApplication];
            if (![app setActivationPolicy:NSApplicationActivationPolicyRegular]) {
                return 0;
            }

            NSString *url = web_url == NULL ? @"" : [NSString stringWithUTF8String:web_url];
            grokBuildSwitchDelegate = [[GrokBuildSwitchAppDelegate alloc] initWithWebURL:url];
            [app setDelegate:grokBuildSwitchDelegate];

            NSString *appName = macapp_application_name();
            NSMenu *mainMenu = [[NSMenu alloc] initWithTitle:@""];
            NSMenuItem *applicationMenuItem = [[NSMenuItem alloc] initWithTitle:@"" action:nil keyEquivalent:@""];
            [mainMenu addItem:applicationMenuItem];

            NSMenu *applicationMenu = [[NSMenu alloc] initWithTitle:appName];
            NSMenuItem *openItem = [[NSMenuItem alloc] initWithTitle:@"打开管理界面"
                                                            action:@selector(openWebInterface:)
                                                     keyEquivalent:@"o"];
            [openItem setTarget:grokBuildSwitchDelegate];
            [applicationMenu addItem:openItem];
            [applicationMenu addItem:[NSMenuItem separatorItem]];

            NSString *quitTitle = [NSString stringWithFormat:@"退出 %@", appName];
            NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:quitTitle
                                                            action:@selector(terminate:)
                                                     keyEquivalent:@"q"];
            [applicationMenu addItem:quitItem];
            [applicationMenuItem setSubmenu:applicationMenu];
            [app setMainMenu:mainMenu];
            return 1;
        } @catch (NSException *exception) {
            fprintf(stderr, "AppKit initialization exception: %s: %s\n",
                    exception.name.UTF8String,
                    exception.reason.UTF8String);
            return -2;
        }
    }
}

void macapp_run(void) {
    @autoreleasepool {
        [NSApp run];
    }
}

void macapp_request_exit(void) {
    [grokBuildSwitchDelegate performSelectorOnMainThread:@selector(requestExit)
                                              withObject:nil
                                           waitUntilDone:NO];
}
