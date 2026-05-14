import UIKit
import UniformTypeIdentifiers

class ShareViewController: UIViewController {
    let APP_GROUP_ID = "group.ink.saveto.app"
    let APP_URL_SCHEME = "savetoink"

    override public func viewDidLoad() {
        super.viewDidLoad()
        print("[ShareExtension] viewDidLoad")

        guard let extensionItem = extensionContext?.inputItems.first as? NSExtensionItem,
              let attachments = extensionItem.attachments else {
            print("[ShareExtension] No extension items or attachments found")
            self.exit()
            return
        }

        print("[ShareExtension] Found \(attachments.count) attachment(s)")

        var texts: [String] = []
        let group = DispatchGroup()

        for (index, provider) in attachments.enumerated() {
            if provider.hasItemConformingToTypeIdentifier("public.url") {
                group.enter()
                provider.loadItem(forTypeIdentifier: "public.url", options: nil) { item, error in
                    if let error {
                        print("[ShareExtension] Failed to load URL attachment \(index): \(error.localizedDescription)")
                    } else if let url = item as? URL {
                        print("[ShareExtension] Got URL: \(url.absoluteString)")
                        texts.append(url.absoluteString)
                    } else {
                        print("[ShareExtension] URL attachment \(index) was not a URL: \(String(describing: item))")
                    }
                    group.leave()
                }
            } else if provider.hasItemConformingToTypeIdentifier("public.plain-text") {
                group.enter()
                provider.loadItem(forTypeIdentifier: "public.plain-text", options: nil) { item, error in
                    if let error {
                        print("[ShareExtension] Failed to load text attachment \(index): \(error.localizedDescription)")
                    } else if let text = item as? String {
                        print("[ShareExtension] Got text: \(text)")
                        texts.append(text)
                    }
                    group.leave()
                }
            } else {
                print("[ShareExtension] Attachment \(index) has unsupported type")
            }
        }

        group.notify(queue: .main) { [weak self] in
            guard let self else { return }

            print("[ShareExtension] Collected \(texts.count) text(s)")

            let shareData: [String: Any] = [
                "title": extensionItem.attributedTitle?.string ?? "",
                "texts": texts,
                "files": []
            ]
            let userDefaults = UserDefaults(suiteName: self.APP_GROUP_ID)
            userDefaults?.set(shareData, forKey: "share-target-data")
            let synced = userDefaults?.synchronize() ?? false
            print("[ShareExtension] Saved to UserDefaults, synced: \(synced)")

            let urlString = "\(self.APP_URL_SCHEME)://share"
            print("[ShareExtension] Attempting to open: \(urlString)")
            if let url = URL(string: urlString) {
                self.openURL(url)
            }
            self.exit()
        }
    }

    private func openURL(_ url: URL) {
        var responder: UIResponder? = self
        while responder != nil {
            if let app = responder as? UIApplication {
                print("[ShareExtension] Found UIApplication via responder chain")
                app.perform(#selector(UIApplication.openURL(_:)), with: url)
                return
            }
            responder = responder?.next
        }
        print("[ShareExtension] ERROR: Could not find UIApplication in responder chain")
    }

    private func exit() {
        print("[ShareExtension] Completing share request")
        extensionContext?.completeRequest(returningItems: nil, completionHandler: nil)
    }
}
