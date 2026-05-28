// Date: 2026-05-28
// Author: XinYang Li

"use client";

import { FileText, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";

import { getStoredToken } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { Attachment } from "@/types/domain";

/**
 * Renders uploaded attachments as image tiles or file chips for either pending or persisted message states.
 * @param props.attachments The ordered attachment list rendered inside the strip.
 * @param props.onRemove Optional remove handler used by the pending composer strip.
 * @param props.compact Whether the strip should use the compact bubble sizing.
 * @returns The attachment strip element.
 */
export function AttachmentStrip({
  attachments,
  compact = false,
  onRemove,
}: {
  attachments: Attachment[];
  compact?: boolean;
  onRemove?: (attachmentId: string) => void;
}): JSX.Element | null {
  const [imagePreviewUrls, setImagePreviewUrls] = useState<Record<string, string>>({});
  const createdPreviewUrlsRef = useRef<Record<string, string>>({});
  const token = useMemo(() => getStoredToken(), []);

  useEffect(() => {
    if (!token) {
      return;
    }

    const imageAttachments = attachments.filter(
      (attachment) => attachment.sourceType === "image" && !createdPreviewUrlsRef.current[attachment.id],
    );
    if (imageAttachments.length === 0) {
      return;
    }

    const controller = new AbortController();

    Promise.all(
      imageAttachments.map(async (attachment) => {
        const response = await fetch(attachment.sourceUrl, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
          signal: controller.signal,
        });
        if (!response.ok) {
          throw new Error(`failed to load image ${attachment.id}`);
        }
        const blob = await response.blob();
        const objectUrl = URL.createObjectURL(blob);
        return [attachment.id, objectUrl] as const;
      }),
    )
      .then((entries) => {
        entries.forEach(([attachmentId, objectUrl]) => {
          createdPreviewUrlsRef.current[attachmentId] = objectUrl;
        });
        setImagePreviewUrls((current) => ({
          ...current,
          ...Object.fromEntries(entries),
        }));
      })
      .catch(() => {
        // Ignore preview failures and keep the fallback tile.
      });

    return () => {
      controller.abort();
    };
  }, [attachments, token]);

  useEffect(() => {
    return () => {
      Object.values(createdPreviewUrlsRef.current).forEach((url) => URL.revokeObjectURL(url));
    };
  }, []);

  if (attachments.length === 0) {
    return null;
  }

  return (
    <div className={cn("space-y-3", compact ? "mb-3" : "mb-4")}>
      <div className="flex flex-wrap gap-3">
        {attachments.map((attachment) => {
          if (attachment.sourceType === "image") {
            return (
              <div className="group relative overflow-hidden rounded-[22px] border border-line bg-white shadow-sm" key={attachment.id}>
                    {imagePreviewUrls[attachment.id] ? (
                      <img
                        alt={attachment.fileName}
                        className={cn("block object-cover", compact ? "h-24 w-24" : "h-28 w-28")}
                        src={imagePreviewUrls[attachment.id]}
                      />
                    ) : (
                      <div className={cn("flex items-center justify-center bg-mist text-ink/35", compact ? "h-24 w-24" : "h-28 w-28")}>
                        图片
                      </div>
                    )}
                <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/55 to-transparent px-3 py-2 text-xs text-white">
                  <p className="truncate">{attachment.fileName}</p>
                </div>
                {onRemove ? (
                  <Button
                    className="absolute right-2 top-2 h-8 w-8 rounded-full bg-white/88 p-0 text-ink hover:bg-white"
                    onClick={() => onRemove(attachment.id)}
                    size="icon"
                    type="button"
                    variant="ghost"
                  >
                    <X className="h-3.5 w-3.5" />
                  </Button>
                ) : null}
              </div>
            );
          }

          return (
            <div
              className={cn(
                "group flex items-center gap-3 rounded-[20px] border border-line bg-white px-4 py-3 shadow-sm",
                compact ? "min-w-[180px] max-w-[240px]" : "min-w-[220px] max-w-[280px]",
              )}
              key={attachment.id}
            >
              <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-mist text-ink/65">
                <FileText className="h-4 w-4" />
              </span>
              <button
                className="min-w-0 flex-1 text-left"
                onClick={() => {
                  if (!token) {
                    return;
                  }

                  fetch(attachment.sourceUrl, {
                    headers: {
                      Authorization: `Bearer ${token}`,
                    },
                  })
                    .then(async (response) => {
                      if (!response.ok) {
                        throw new Error("download failed");
                      }
                      const blob = await response.blob();
                      const objectUrl = URL.createObjectURL(blob);
                      const anchor = document.createElement("a");
                      anchor.href = objectUrl;
                      anchor.download = attachment.fileName;
                      anchor.click();
                      window.setTimeout(() => URL.revokeObjectURL(objectUrl), 1000);
                    })
                    .catch(() => {
                      // Ignore download failures inside the visual strip.
                    });
                }}
                type="button"
              >
                <p className="truncate text-sm text-ink">{attachment.fileName}</p>
              </button>
              {onRemove ? (
                <Button
                  className="h-8 w-8 rounded-full p-0 text-ink/52 hover:text-ink"
                  onClick={() => onRemove(attachment.id)}
                  size="icon"
                  type="button"
                  variant="ghost"
                >
                  <X className="h-3.5 w-3.5" />
                </Button>
              ) : null}
            </div>
          );
        })}
      </div>
    </div>
  );
}
