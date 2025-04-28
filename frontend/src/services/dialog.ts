import {type App, type ComponentPublicInstance, createApp} from 'vue'
import FullscreenModal from "@/components/FullscreenModal.vue";

type ModalInstance = ComponentPublicInstance & {
  open: (text: string) => void
  close: () => void
}

class DialogService {
  private static instance: DialogService
  private modalInstance: ModalInstance | null = null
  private app: App | null = null

  private constructor() {}

  public static getInstance(): DialogService {
    if (!DialogService.instance) {
      DialogService.instance = new DialogService()
    }
    return DialogService.instance
  }

  private ensureModal(): void {
    if (!this.modalInstance) {
      const modalContainer = document.createElement('div')
      modalContainer.id = 'modal-container'
      document.body.appendChild(modalContainer)

      this.app = createApp(FullscreenModal)
      this.modalInstance = this.app.mount(modalContainer) as ModalInstance
    }
  }

  public show(text: string): void {
    this.ensureModal()
    this.modalInstance?.open(text)
  }

  public close(): void {
    if (this.modalInstance) {
      this.modalInstance.close()
    }
  }

  public destroy(): void {
    if (this.app) {
      this.app.unmount()
      const container = document.getElementById('modal-container')
      if (container) {
        document.body.removeChild(container)
      }
      this.modalInstance = null
      this.app = null
    }
  }
}

export const Dialog = DialogService.getInstance()

if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', () => {
    Dialog.destroy()
  })
}
