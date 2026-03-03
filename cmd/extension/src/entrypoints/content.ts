import { mount } from "svelte";
import Toast from "./content/Toast.svelte";

export default defineContentScript({
    matches: ["<all_urls>"],
    main() {
        let toastCount = 0;

        function createToast(title: string, message: string, variant: "success" | "error" = "success") {
            const containerId = `savetoink-toast-${toastCount++}`;
            const container = document.createElement("div");
            container.id = containerId;
            document.body.appendChild(container);

            const toast = mount(Toast, {
                target: container,
                props: {
                    title,
                    message,
                    variant,
                },
            });

            return toast;
        }

        function showToast(message: { type: string; title: string; message: string; variant?: "success" | "error" }) {
            if (message.type !== "show-toast") {
                return;
            }

            createToast(message.title, message.message, message.variant || "success");
        }

        browser.runtime.onMessage.addListener((message) => {
            showToast(message);
        });
    },
});
