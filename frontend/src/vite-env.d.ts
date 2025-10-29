/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string
  // aquí puedes agregar más variables públicas si luego creas más en tu .env
  // readonly VITE_OTRA_COSA: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}