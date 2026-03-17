# Mobile Application Development Plan

## Overview

Develop a native mobile application for SaveToInk using [CapacitorJS](https://capacitorjs.com/). The app is essentially a fullscreen WebView wrapper around the live webapp, with a native share receiver that intercepts shared URLs from other apps and opens the webapp to `/new?url=<url>`. The webapp URL is injected at build time via the `PUBLIC_APP_URL` environment variable.

**Note**: The Capacitor Share API ([@capacitor/share](https://capacitorjs.com/docs/apis/share)) is for **outgoing** shares (sharing FROM your app TO other apps). For **incoming** shares (receiving URLs FROM other apps), we need minimal platform-specific native code.

## Goals

- **Minimal Native Code**: App is 99% web, only share receiver is native
- **Zero Auth Complexity**: Uses existing Auth0 session in WebView (cookies persist)
- **Full Feature Parity**: All webapp features work automatically
- **Incoming Share**: Users can share links from any app directly to SaveToInk
- **Simple Maintenance**: Changes to webapp immediately available in mobile app

## Technology Stack

- **Package Manager**: Bun
- **Framework**: CapacitorJS 6.x
- **UI**: Live webapp URL injected via `PUBLIC_APP_URL` build-time environment variable
- **Authentication**: Existing Auth0 session (cookies persist in WebView)
- **Incoming Share**: Minimal native code for Android intents and iOS share extension
- **Platforms**: iOS, Android

---

## 1. Project Setup - COMPLETED

### 1.1 Initialize Capacitor Project

```bash
# Create new Capacitor app in frontend/mobile
cd frontend
mkdir mobile && cd mobile
bun init -y

# Install Capacitor core
bun install @capacitor/core @capacitor/cli @capacitor/android @capacitor/ios
bunx cap init SaveToInk ink.saveto.app
```

### 1.2 Configure Capacitor to Load Live Site

```typescript
// capacitor.config.ts
import { defineConfig } from '@capacitor/cli';

// URL comes from build-time env var, no fallback in code
const MOBILE_DEV_URL = process.env.MOBILE_DEV_URL || 'http://localhost:4000';

export default defineConfig({
  appId: 'ink.saveto.app',
  appName: 'SaveToInk',
  webDir: 'dist',
  bundledWebRuntime: false,
  server: {
    // For development, can point to local webapp
    url: MOBILE_DEV_URL,
    cleartext: false,
    androidScheme: 'https'
  },
  plugins: {
    // No plugins needed for basic WebView wrapping
  }
});
```

### 1.2 Configure Capacitor to Load Live Site

```typescript
// capacitor.config.ts
import { defineConfig } from '@capacitor/cli';

export default defineConfig({
  appId: 'ink.saveto.app',
  appName: 'Save to Ink',
  webDir: 'dist',  // Can be empty - we'll redirect to live site
  bundledWebRuntime: false,
  server: {
    // For development, can point to local webapp
    url: process.env.MOBILE_DEV_URL || 'http://localhost:4000',
    cleartext: false,
    androidScheme: 'https'
  },
  plugins: {
    // No plugins needed for basic WebView wrapping
  }
});
```

### 1.3 Create Minimal Web Assets

```html
<!-- src/index.html -->
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no" />
    <title>Save to Ink</title>
  </head>
  <body>
    <script type="module">
      // Vite injects PUBLIC_APP_URL at build time - fail if missing
      const APP_URL = import.meta.env.PUBLIC_APP_URL;

      if (!APP_URL) {
        document.body.innerHTML = '<h1>Error: PUBLIC_APP_URL not set at build time</h1>';
        console.error('PUBLIC_APP_URL is not defined at build time');
        throw new Error('PUBLIC_APP_URL must be set at build time');
      }

      // Handle shared URLs from native share extension
      const urlParams = new URLSearchParams(window.location.search);
      const sharedUrl = urlParams.get('shared_url');

      if (sharedUrl) {
        // Navigate to /new with shared URL
        window.location.href = `${APP_URL}/new?url=${encodeURIComponent(sharedUrl)}`;
      } else {
        // Navigate to home
        window.location.href = APP_URL;
      }
    </script>
  </body>
</html>
```

### 1.4 Vite Configuration

```typescript
// vite.config.ts
import { defineConfig } from 'vite';

export default defineConfig({
  base: './',
  build: {
    target: 'esnext',
    outDir: 'dist',
    emptyOutDir: true
  }
});
```

---

## 2. Android Incoming Share Intent

**Note**: Capacitor's Share API ([@capacitor/share](https://capacitorjs.com/docs/apis/share)) is for **outgoing** shares (sharing FROM your app TO other apps). To receive **incoming** shares (receiving URLs FROM other apps), we need minimal platform-specific native code. This is standard Android development.

### 2.1 Configure Intent Filter

Edit `android/app/src/main/AndroidManifest.xml`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android">
    <application
        android:allowBackup="true"
        android:icon="@mipmap/ic_launcher"
        android:label="@string/app_name"
        android:theme="@style/AppTheme">

        <activity
            android:name=".MainActivity"
            android:exported="true"
            android:theme="@style/AppTheme.NoActionBarLaunch">

            <!-- Main app launcher -->
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>

            <!-- Handle share intent (text/uri-list) -->
            <intent-filter android:label="Save to Ink">
                <action android:name="android.intent.action.SEND" />
                <category android:name="android.intent.category.DEFAULT" />
                <data android:mimeType="text/uri-list" />
            </intent-filter>

            <!-- Handle share intent (text/plain) -->
            <intent-filter android:label="Save to Ink">
                <action android:name="android.intent.action.SEND" />
                <category android:name="android.intent.category.DEFAULT" />
                <data android:mimeType="text/plain" />
            </intent-filter>

            <!-- Handle multiple share intents -->
            <intent-filter android:label="Save to Ink">
                <action android:name="android.intent.action.SEND_MULTIPLE" />
                <category android:name="android.intent.category.DEFAULT" />
                <data android:mimeType="text/plain" />
            </intent-filter>
        </activity>
    </application>
</manifest>
```

### 2.2 Add Build-Time URL Configuration

First, add the app URL as a build-time configuration in `android/app/build.gradle`:

```gradle
android {
  // ...

  defaultConfig {
    // ...
    buildConfigField "String", "APP_URL", project.hasProperty("APP_URL") ? project.property("APP_URL") : "\"\""
  }
}
```

Then edit `android/app/src/main/java/ink/saveto/app/MainActivity.java`:

```java
package ink.saveto.app;

import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import com.getcapacitor.BridgeActivity;

public class MainActivity extends BridgeActivity {

  @Override
  public void onCreate(Bundle savedInstanceState) {
    super.onCreate(savedInstanceState);

    // Handle incoming share intent
    handleShareIntent(getIntent());
  }

  @Override
  public void onNewIntent(Intent intent) {
    super.onNewIntent(intent);
    handleShareIntent(intent);
  }

  private void handleShareIntent(Intent intent) {
    if (intent == null) {
      return;
    }

    String action = intent.getAction();
    String type = intent.getType();

    // Check if this is a share intent
    if (Intent.ACTION_SEND.equals(action) && type != null) {
      // Get shared URL
      String sharedUrl = null;

      if ("text/uri-list".equals(type) || "text/plain".equals(type)) {
        Uri uri = intent.getParcelableExtra(Intent.EXTRA_STREAM);
        if (uri != null) {
          sharedUrl = uri.toString();
        } else {
          sharedUrl = intent.getStringExtra(Intent.EXTRA_TEXT);
        }
      }

      // Navigate WebView to /new with shared URL using build-time APP_URL
      if (sharedUrl != null) {
        String encodedUrl = Uri.encode(sharedUrl);
        String targetUrl = BuildConfig.APP_URL + "/new?url=" + encodedUrl;

        // Load URL in WebView
        if (getBridge() != null && getBridge().getWebView() != null) {
          getBridge().getWebView().loadUrl(targetUrl);
        }
      }
    }
  }
}
```

---

## 3. iOS Share Extension

**Note**: iOS requires a Share Extension target to receive shares from other apps. This is standard iOS development using native Swift. The extension simply passes the shared URL to the main app via a custom URL scheme.

### 3.1 Create Share Extension

In Xcode:
1. File → New → Target → Share Extension
2. Name it "SaveToInkShareExtension"
3. Configure the extension

### 3.2 Share Extension Code

```swift
// SaveToInkShareExtension/ShareViewController.swift
import UIKit
import Social
import MobileCoreServices

class ShareViewController: UIViewController {

  override func viewDidLoad() {
    super.viewDidLoad()

    // Get shared content
    guard let extensionItem = extensionContext?.inputItems.first as? NSExtensionItem,
          let itemProvider = extensionItem.attachments?.first,
          itemProvider.hasItemConformingToTypeIdentifier("public.url") else {
      self.cancel()
      return
    }

    // Load URL
    itemProvider.loadItem(forTypeIdentifier: "public.url", options: nil) { [weak self] (item, error) in
      guard let url = item as? URL else {
        self?.cancel()
        return
      }

      // Open main app with URL
      self?.openMainApp(with: url.absoluteString)
    }
  }

  private func openMainApp(with urlString: String) {
    // Create URL to open main app with shared URL
    let urlString = "savetoink://new?url=\(urlString.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? urlString)"

    // Use responder chain to open URL
    var responder: UIResponder? = self
    while responder != nil {
      if let application = responder as? UIApplication {
        application.perform(#selector(openURL(_:)), with: URL(string: urlString))
        break
      }
      responder = responder?.next
    }

    // Complete extension
    self.extensionContext?.completeRequest(returningItems: nil, completionHandler: nil)
  }

  private func cancel() {
    self.extensionContext?.completeRequest(returningItems: nil, completionHandler: nil)
  }
}
```

### 3.3 Add Build-Time URL Configuration

Add the app URL as a build-time configuration in `ios/App/App/Info.plist`:

```xml
<!-- Add inside the root <dict> element -->
<key>APP_URL</key>
<string>$(APP_URL)</string>
```

Then configure the build in Xcode:
1. Select the App target
2. Go to Build Settings
3. Add User-Defined Setting: `APP_URL`
4. Set the `APP_URL` User-Defined Setting to your production webapp URL

### 3.4 Configure Main App URL Scheme

Edit `ios/App/App/Info.plist`:

```xml
<key>CFBundleURLTypes</key>
<array>
  <dict>
    <key>CFBundleURLName</key>
    <string>ink.saveto.app</string>
    <key>CFBundleURLSchemes</key>
    <array>
      <string>savetoink</string>
    </array>
    <key>CFBundleURLTypes</key>
    <string>Editor View</string>
  </dict>
</array>
<!-- Add app URL for build-time injection -->
<key>APP_URL</key>
<string>$(APP_URL)</string>
```

### 3.5 Handle URL Scheme in Main App

```swift
// ios/App/App/App.swift
import UIKit
import Capacitor

@UIApplicationMain
class AppDelegate: UIResponder, UIApplicationDelegate {

  var window: UIWindow?

  func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
    return true
  }

  func application(_ app: UIApplication, open url: URL, options: [UIApplication.OpenURLOptionsKey : Any] = [:]) -> Bool {
    // Handle deep link from share extension
    if url.scheme == "savetoink" {
      // Extract URL and navigate WebView using build-time APP_URL
      let urlString = url.absoluteString
      guard let appUrl = Bundle.main.object(forInfoDictionaryKey: "APP_URL") as? String, !appUrl.isEmpty else {
        print("ERROR: APP_URL not set in Info.plist")
        return false
      }
      let webappUrl = "\(appUrl)/new?\(urlString.replacingOccurrences(of: "savetoink://new?url=", with: ""))"

      if let bridge = (window?.rootViewController as? CAPBridgeViewController)?.bridge {
        bridge.getWebView().load(URLRequest(url: URL(string: webappUrl)!))
      }
    }

    return ApplicationDelegateProxy.shared.application(app, open: url, options: options)
  }

  func application(_ application: UIApplication, continue userActivity: NSUserActivity, restorationHandler: @escaping ([UIUserActivityRestoring]?) -> Void) -> Bool {
    return true
  }
}
```



---

## 4. Development Workflow

### 4.1 Development Setup

```bash
# For development, you can point to local webapp
cd frontend/mobile
MOBILE_DEV_URL=http://localhost:4000 bun run dev

# In another terminal, start webapp
cd frontend/webapp
bun run dev
```

### 4.2 Sync and Run on Device

```bash
# Sync native platforms
cd frontend/mobile
bunx cap sync

# Open in Xcode
bunx cap open ios

# Open in Android Studio
bunx cap open android
```

### 4.3 Development Flow

1. Make changes to webapp in `frontend/webapp`
2. Webapp dev server auto-reloads
3. If testing on emulator/simulator, refresh WebView (changes live-reload if webapp supports it)
4. For share extension testing, rebuild and reinstall the app

### 4.4 Production Build

```bash
# Set production URL and build web assets
cd frontend/mobile
# Set environment variable before running
export APP_URL=https://your-app-url.com
PUBLIC_APP_URL=$APP_URL bun run build

# Sync to native platforms
bunx cap sync

# Build release versions
# iOS: Set APP_URL in Xcode User-Defined Settings, then archive
# Android: Set APP_URL gradle property: APP_URL=$APP_URL ./gradlew assembleRelease
```

---

## 5. Testing

### 5.1 Test Share Functionality

**Android:**
1. Open any app (browser, social media, etc.)
2. Find a link
3. Tap share button
4. Select "Save to Ink" from share sheet
5. Verify app opens and navigates to `/new` with the URL pre-filled

**iOS:**
1. Open any app (Safari, social media, etc.)
2. Find a link
3. Tap share button
4. Swipe to find "Save to Ink" in share sheet
5. Tap and verify app opens with URL pre-filled

### 5.2 Test WebView Functionality

- Verify all webapp features work (login, articles list, article detail, etc.)
- Test Auth0 login/logout flow
- Verify session persists across app restarts
- Test responsive design on mobile viewport

### 5.3 Test Edge Cases

- Share invalid URLs
- Share non-URL text
- Share multiple URLs (Android SEND_MULTIPLE)
- App already open when sharing
- Network errors
- Deep linking from other apps

---

## 6. Configuration

### 6.1 Environment Variables

```bash
# .env.local (development)
MOBILE_DEV_URL=http://localhost:4000

# Production build (injects at build time)
# Set environment variable before running this (replace with your actual URL)
# export APP_URL=https://your-webapp-domain.com
PUBLIC_APP_URL=$APP_URL
```

### 6.2 App Metadata

**Android** (`android/app/src/main/res/values/strings.xml`):

```xml
<resources>
  <string name="app_name">Save to Ink</string>
</resources>
```

**iOS** (`ios/App/App/Info.plist`):

```xml
<key>CFBundleDisplayName</key>
<string>Save to Ink</string>
```

### 6.3 App Icons

Generate app icons using existing icon generation tools:

```bash
cd frontend/mobile
bun run icons:generate  # Reuse existing icon generation
```

### 6.4 WebView Configuration

Configure WebView behavior. The `PUBLIC_APP_URL` is injected at build time:

```typescript
// capacitor.config.ts
const APP_URL = import.meta.env.PUBLIC_APP_URL;

export default defineConfig({
  // ...
  server: {
    androidScheme: 'https',
    cleartext: false,
    allowNavigation: [APP_URL, 'https://*.auth0.com']  // Allow Auth0 (app URL comes from env var)
  }
});
```

---

## 7. Deployment

### 7.1 App Store Submission

**iOS:**
1. Configure app signing in Xcode
2. Archive and upload to App Store Connect
3. Add app store screenshots and description
4. Submit for review

**Android:**
1. Generate signed APK/AAB
2. Upload to Google Play Console
3. Add store listing and screenshots
4. Submit for review

### 7.2 CI/CD Pipeline

```yaml
# .github/workflows/mobile-build.yml
name: Build Mobile App

on:
  push:
    branches: [main]
    paths: ['frontend/mobile/**']

jobs:
  build-android:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Bun
        uses: oven-sh/setup-bun@v1
        with:
          bun-version: latest
      - name: Install dependencies
        run: |
          cd frontend/mobile
          bun install
      - name: Build web
        run: |
          cd frontend/mobile
          bun run build
      - name: Sync Android
        run: |
          cd frontend/mobile
          bunx cap sync android
      - name: Build Android
        run: |
          cd frontend/mobile/android
          ./gradlew assembleRelease

  build-ios:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Bun
        uses: oven-sh/setup-bun@v1
        with:
          bun-version: latest
      - name: Install dependencies
        run: |
          cd frontend/mobile
          bun install
      - name: Build web
        run: |
          cd frontend/mobile
          bun run build
      - name: Sync iOS
        run: |
          cd frontend/mobile
          bunx cap sync ios
      - name: Build iOS
        run: |
          cd frontend/mobile/ios
          xcodebuild -workspace App.xcworkspace -scheme App -configuration Release archive
```

---

## 8. Development Tasks Checklist

### Phase 1: Basic Capacitor Setup (Week 1, Day 1-2)
- [ ] Initialize Capacitor project in `frontend/mobile`
- [ ] Configure capacitor.config.ts to load live site
- [ ] Create minimal index.html with PUBLIC_APP_URL env var injection
- [ ] Set up Vite configuration
- [ ] Test basic WebView loads the configured webapp URL (injected from PUBLIC_APP_URL)

### Phase 2: Android Share Intent (Week 1, Day 3-4)
- [ ] Configure AndroidManifest.xml intent filters
- [ ] Implement handleShareIntent in MainActivity.java
- [ ] Test share from Android browser and other apps
- [ ] Fix any issues with URL parsing

### Phase 3: iOS Share Extension (Week 1, Day 5)
- [ ] Create iOS share extension target in Xcode
- [ ] Implement ShareViewController.swift
- [ ] Configure URL scheme in Info.plist
- [ ] Implement URL handling in AppDelegate
- [ ] Test share from iOS Safari and other apps

### Phase 4: Polish and Testing (Week 2)
- [ ] Test on multiple Android devices/OS versions
- [ ] Test on multiple iOS devices/OS versions
- [ ] Test edge cases (invalid URLs, network errors, etc.)
- [ ] Configure app icons and metadata
- [ ] Test full user journey (login → share → save)

### Phase 5: Build & Deploy (Week 2)
- [ ] Configure code signing for both platforms
- [ ] Build release versions
- [ ] Create app store listings
- [ ] Prepare screenshots and descriptions
- [ ] Set up CI/CD pipeline
- [ ] Submit to TestFlight/internal testing
- [ ] Submit to app stores

---

## 9. Future Enhancements

### Potential Features
- Offline support (service worker)
- Push notifications
- Biometric authentication for app access
- Custom app theming
- Background sync

### Improvements
- Add "Open in Browser" option
- Add in-app browser for external links
- Better error handling for share failures
- Analytics tracking
- Crash reporting

---

## 10. Risks & Mitigation

### Risk 1: WebView performance issues
- **Mitigation**: Test on low-end devices
- **Mitigation**: Ensure webapp is optimized for performance
- **Mitigation**: Consider enabling WebView caching

### Risk 2: Share extension complexity on iOS
- **Mitigation**: Use the minimal implementation provided (~50 lines of Swift)
- **Mitigation**: Test extensively on various iOS versions
- **Mitigation**: Follow Apple's Share Extension documentation closely

### Risk 3: Auth0 session issues in WebView
- **Mitigation**: Test login/logout flow thoroughly
- **Mitigation**: Ensure cookies persist correctly
- **Mitigation**: Configure proper cookie settings in webapp

### Risk 4: App Store rejection
- **Mitigation**: Follow platform guidelines
- **Mitigation**: Submit to TestFlight early
- **Mitigation**: Ensure app provides clear value

---

## 11. Success Metrics

- **Functionality**: Share works from any app on both platforms
- **User Experience**: App loads and performs as well as browser
- **Feature Parity**: All webapp features work in app
- **User Satisfaction**: Store rating >4.5/5 after 100 reviews
- **Adoption**: >1000 downloads in first month

---

## 12. Comparison with Alternatives

### Why Not Use @capacitor/share?

The Capacitor Share API ([@capacitor/share](https://capacitorjs.com/docs/apis/share)) is designed for **outgoing** shares (sharing content FROM your app TO other apps). For **incoming** shares (receiving URLs FROM other apps), native platform-specific code is required:

- **Android**: Intent filters + MainActivity (~30 lines of code)
- **iOS**: Share Extension target + URL scheme (~50 lines of Swift)

This is minimal, standard platform development and requires no custom Capacitor plugins.

### vs. Full Native UI
| Aspect | WebView Approach | Native UI |
|--------|------------------|-----------|
| Development Time | 1-2 weeks | 6-8 weeks |
| Code Maintenance | Webapp changes auto-apply | Need to update native code |
| Feature Parity | 100% (uses live site) | Need to recreate features |
| Authentication | Uses existing Auth0 | Need native auth handling |
| Token Storage | Cookies (handled by browser) | Need native storage |
| Incoming Share | Minimal native code (~80 lines total) | Entire app is native |
| App Size | Small (web assets only) | Larger (includes all code) |
| Performance | Slightly slower (WebView) | Faster (native) |
| URL Configuration | Via build-time env var | Via build-time config |

**Recommendation**: WebView approach is ideal for this use case because:
- Time to market is dramatically faster
- Maintenance is trivial (webapp updates = app updates)
- Only ~80 lines of native code needed total (Android + iOS)
- Performance difference is negligible for this use case
- Auth0 integration is handled automatically
- URL configuration is centralized in build-time env vars

---

## 13. References

- [Capacitor Documentation](https://capacitorjs.com/docs)
- [Capacitor Share API](https://capacitorjs.com/docs/apis/share) - Note: Used for outgoing shares only
- [Capacitor Deep Linking](https://capacitorjs.com/docs/guides/deep-links)
- [Android Share Intents](https://developer.android.com/training/sharing/receive)
- [iOS Share Extensions](https://developer.apple.com/documentation/uikit/uiactivityviewcontroller)
- [Android Intent Filters](https://developer.android.com/guide/topics/manifest/intent-filter-element)
- [iOS URL Schemes](https://developer.apple.com/documentation/xcode/defining-a-custom-url-scheme-for-your-app)
- [Environment Variables in Capacitor](https://capacitorjs.com/docs/guides/environment-variables)
- [Gradle BuildConfig Fields](https://developer.android.com/build/config/build-variants#buildconfig)
