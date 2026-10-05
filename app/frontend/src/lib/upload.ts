import { postApiV1StoragePresigned } from "@/api/endpoints/storage/storage"

export async function uploadToR2(file: File, folder: string = "products"): Promise<string> {
  try {
    const res = await postApiV1StoragePresigned({
      file_name: `${folder}/${file.name}`,
      content_type: file.type || "application/octet-stream",
    })

    const payload = (res as any)?.data || res
    const uploadUrl = payload?.upload_url
    const publicUrl = payload?.public_url

    if (!uploadUrl || !publicUrl) {
      throw new Error("Invalid presigned URL response from storage service")
    }

    const uploadRes = await fetch(uploadUrl, {
      method: "PUT",
      headers: {
        "Content-Type": file.type || "application/octet-stream",
      },
      body: file,
    })

    if (!uploadRes.ok) {
      throw new Error(`Failed to upload file to storage: ${uploadRes.status} ${uploadRes.statusText}`)
    }

    return publicUrl
  } catch (error) {
    console.error("[uploadToR2] Error:", error)
    throw error
  }
}
