<script lang="ts">
    import { onMount } from "svelte";
    
    export let title: string;
    export let message: string;
    export let variant: "success" | "error" = "success";
    
    let container: HTMLDivElement;
    let timeoutId: NodeJS.Timeout;
    
    onMount(() => {
        timeoutId = setTimeout(() => {
            dismiss();
        }, 4000);
        
        return () => {
            if (timeoutId) clearTimeout(timeoutId);
        };
    });
    
    function dismiss() {
        container.style.opacity = "0";
        setTimeout(() => {
            if (container.parentNode) {
                container.parentNode.removeChild(container);
            }
        }, 300);
    }
</script>

<div bind:this={container} class="toast toast-{variant}">
    <div class="toast-content">
        <strong>{title}</strong>
        <span>{message}</span>
    </div>
    <button class="toast-close" on:click={dismiss}>×</button>
</div>

<style>
    .toast {
        position: fixed;
        bottom: 20px;
        right: 20px;
        z-index: 999999;
        min-width: 300px;
        max-width: 400px;
        background: var(--card-background-color, #fff);
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        padding: 12px 16px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        transition: opacity 0.3s ease;
        animation: slideIn 0.3s ease;
    }
    
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    
    .toast-success {
        border-left: 4px solid #198754;
    }
    
    .toast-error {
        border-left: 4px solid #dc3545;
    }
    
    .toast-content {
        display: flex;
        flex-direction: column;
        gap: 4px;
        flex: 1;
    }
    
    .toast-content strong {
        font-size: 14px;
        font-weight: 600;
    }
    
    .toast-content span {
        font-size: 13px;
        color: var(--muted-color, #666);
    }
    
    .toast-close {
        background: none;
        border: none;
        font-size: 20px;
        line-height: 1;
        cursor: pointer;
        padding: 0;
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--muted-color, #666);
        opacity: 0.6;
        transition: opacity 0.2s;
    }
    
    .toast-close:hover {
        opacity: 1;
    }
</style>
