import UIKit

class SceneDelegate: UIResponder, UISceneDelegate {

    var window: UIWindow?
    private var didHandleInitialRedirect = false

    func scene(_ scene: UIScene, willConnectTo session: UISceneSession, options connectionOptions: UIScene.ConnectionOptions) {
        // Handle URL if app was launched/activated via URL scheme (e.g. savetoink://share)
        if let urlContext = connectionOptions.urlContexts.first {
            didHandleInitialRedirect = true
            handleURL(urlContext.url)
        }
    }

    func scene(_ scene: UIScene, openURLContexts URLContexts: Set<UIOpenURLContext>) {
        // Handle URL when app is already running
        didHandleInitialRedirect = true
        if let urlContext = URLContexts.first {
            handleURL(urlContext.url)
        }
    }

    func sceneDidBecomeActive(_ scene: UIScene) {
        // On cold start (tap app icon), redirect to APP_URL immediately
        guard !didHandleInitialRedirect else { return }
        didHandleInitialRedirect = true
        openSafari(path: "")
    }

    // MARK: - Share handling

    private func handleURL(_ url: URL) {
        if url.scheme == "savetoink" {
            handleSharedContent()
        }
    }

    private func handleSharedContent() {
        let appGroup = "group.ink.saveto.app"
        let userDefaults = UserDefaults(suiteName: appGroup)

        guard let shareData = userDefaults?.dictionary(forKey: "share-target-data"),
              let texts = shareData["texts"] as? [String],
              let sharedUrl = texts.first,
              sharedUrl.hasPrefix("http") else {
            print("[SceneDelegate] no valid shared URL found, opening webapp home")
            openSafari(path: "")
            return
        }

        // Clear consumed data
        userDefaults?.removeObject(forKey: "share-target-data")

        let encoded = sharedUrl.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? sharedUrl
        openSafari(path: "/new?url=\(encoded)")
    }

    private func openSafari(path: String) {
        guard let appUrl = Bundle.main.object(forInfoDictionaryKey: "APP_URL") as? String,
              !appUrl.isEmpty else {
            print("[SceneDelegate] error: APP_URL not configured in Info.plist")
            return
        }
        let fullUrl = "\(appUrl)\(path)"
        guard let url = URL(string: fullUrl) else {
            print("[SceneDelegate] error: invalid URL: \(fullUrl)")
            return
        }
        UIApplication.shared.open(url, options: [:], completionHandler: nil)
    }
}
