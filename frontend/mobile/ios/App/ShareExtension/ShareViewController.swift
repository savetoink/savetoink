import UIKit
import UniformTypeIdentifiers

class ShareViewController: UIViewController {
    let APP_GROUP_ID = "group.ink.saveto.app"
    let APP_URL_SCHEME = "savetoink"

    override public func viewDidLoad() {
        super.viewDidLoad()

        guard let extensionItem = extensionContext?.inputItems.first as? NSExtensionItem,
                let attachments = extensionItem.attachments else {
            self.exit()
            return
        }

        var texts: [String] = []
        let group = DispatchGroup()

        for provider in attachments {
            // Try URL first
            if provider.hasItemConformingToTypeIdentifier("public.url") {
                group.enter()
                provider.loadItem(forTypeIdentifier: "public.url", options: nil) { item, _ in
                    if let url = item as? URL {
                        texts.append(url.absoluteString)
                    }
                    group.leave()
                }
            } else if provider.hasItemConformingToTypeIdentifier("public.plain-text") {
                group.enter()
                provider.loadItem(forTypeIdentifier: "public.plain-text", options: nil) { item, _ in
                    if let text = item as? String {
                        texts.append(text)
                    }
                    group.leave()
                }
            }
        }

        group.notify(queue: .main) { [weak self] in
            guard let self else { return }
            let shareData: [String: Any] = [
                "title": extensionItem.attributedTitle?.string ?? "",
                "texts": texts,
                "files": []
            ]
            let userDefaults = UserDefaults(suiteName: self.APP_GROUP_ID)
            userDefaults?.set(shareData, forKey: "share-target-data")
            userDefaults?.synchronize()

            // Open main app
            if let url = URL(string: "\(self.APP_URL_SCHEME)://share") {
                self.openURL(url)
            }
            self.exit()
        }
    }

    private func openURL(_ url: URL) {
        var responder: UIResponder? = self
        while responder != nil {
            if let app = responder as? UIApplication {
                app.perform(#selector(UIApplication.openURL(_:)), with: url)
                break
            }
            responder = responder?.next
        }
    }

    private func exit() {
        extensionContext?.completeRequest(returningItems: nil, completionHandler: nil)
    }
}
