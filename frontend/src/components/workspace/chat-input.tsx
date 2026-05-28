// Date: 2026-05-25
// Author: XinYang Li

"use client";

import { ArrowUp, ImagePlus, Paperclip, Plus } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { AttachmentStrip } from "@/components/workspace/attachment-strip";
import { MarkdownContent } from "@/components/workspace/markdown-content";
import { Textarea } from "@/components/ui/textarea";
import type { Attachment, AttachmentSourceType } from "@/types/domain";

/**
 * Renders the fixed chat composer for the task workspace.
 * @param props.placeholder The prompt hint shown inside the textarea.
 * @param props.isSending Whether the current request is already in flight.
 * @param props.isUploadingAttachments Whether one attachment upload request is already in flight.
 * @param props.operationHint Optional active quote or reply context shown above the composer.
 * @param props.pendingAttachments The uploaded attachments that will bind to the next user message.
 * @param props.onClearOperation Clears the active quote or reply context.
 * @param props.onRemoveAttachment Removes one pending attachment from the composer.
 * @param props.onUploadAttachments Uploads a batch of selected files for the active session.
 * @param props.onSend The callback executed after the user submits a new message.
 * @returns The interactive chat input block.
 */
export function ChatInput({
  placeholder,
  isSending,
  isUploadingAttachments,
  operationHint,
  pendingAttachments,
  onClearOperation,
  onRemoveAttachment,
  onUploadAttachments,
  onSend,
}: {
  placeholder: string;
  isSending: boolean;
  isUploadingAttachments: boolean;
  operationHint?: {
    mode: "quote" | "reply";
    sourceContents: string[];
  } | null;
  pendingAttachments: Attachment[];
  onClearOperation?: () => void;
  onRemoveAttachment: (attachmentId: string) => void;
  onUploadAttachments: (sourceType: AttachmentSourceType, files: File[]) => void;
  onSend: (value: string) => void;
}): JSX.Element {
  const [value, setValue] = useState("");
  const [isComposing, setIsComposing] = useState(false);
  const [isUploadMenuOpen, setIsUploadMenuOpen] = useState(false);
  const [pendingEnter, setPendingEnter] = useState(false);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const imageInputRef = useRef<HTMLInputElement | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);

  useEffect(() => {
    const textarea = textareaRef.current;
    if (!textarea) {
      return;
    }

    textarea.style.height = "0px";
    textarea.style.height = `${Math.min(textarea.scrollHeight, 192)}px`;
  }, [value]);

  useEffect(() => {
    if (!pendingEnter) {
      return;
    }

    /**
     * Clears the pending double-enter window when the second Enter key does not arrive in time.
     * @param none The timer does not accept external parameters.
     */
    const timer = window.setTimeout(() => {
      setPendingEnter(false);
    }, 300);

    return () => window.clearTimeout(timer);
  }, [pendingEnter]);

  return (
    <form
      className="rounded-[30px] border border-line bg-white p-4 shadow-panel"
      onSubmit={(event) => {
        event.preventDefault();

        const trimmed = value.trim();
        if (!trimmed || isSending || isUploadingAttachments) {
          return;
        }

        onSend(trimmed);
        setValue("");
        setIsUploadMenuOpen(false);
      }}
    >
      {pendingAttachments.length > 0 ? (
        <AttachmentStrip attachments={pendingAttachments} onRemove={onRemoveAttachment} />
      ) : null}
      {operationHint ? (
        <div className="mb-4 rounded-[22px] border border-line bg-mist/70 px-4 py-3">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="text-xs uppercase tracking-[0.16em] text-pine">
                {operationHint.mode === "quote" ? `引用消息 ${operationHint.sourceContents.length}` : "回复当前消息"}
              </p>
              <div className="mt-2 space-y-2">
                {operationHint.sourceContents.map((sourceContent, index) => (
                  <div className="rounded-2xl border border-line/70 bg-white/70 px-3 py-2" key={`${sourceContent}-${index}`}>
                    <p className="mb-1 text-[11px] uppercase tracking-[0.14em] text-ink/42">
                      {operationHint.mode === "quote" ? `引用消息 ${index + 1}` : "回复消息"}
                    </p>
                    <MarkdownContent
                      content={sourceContent}
                      inlineCodeClassName="rounded-md bg-ink/6 px-1.5 py-0.5 font-mono text-[13px] text-ink"
                      paragraphClassName="whitespace-pre-wrap text-sm leading-6 text-ink/62"
                      tableClassName="overflow-x-auto rounded-2xl border border-line/70 bg-white/75"
                    />
                  </div>
                ))}
              </div>
            </div>
            {onClearOperation ? (
              <Button onClick={onClearOperation} size="sm" type="button" variant="ghost">
                取消
              </Button>
            ) : null}
          </div>
        </div>
      ) : null}
      <Textarea
        className="h-11 min-h-0 max-h-48 resize-none overflow-y-auto border-none p-0 leading-7 focus:ring-0"
        onChange={(event) => setValue(event.target.value)}
        onCompositionEnd={() => setIsComposing(false)}
        onCompositionStart={() => setIsComposing(true)}
        onKeyDown={(event) => {
          if (event.key !== "Enter" || event.shiftKey || event.ctrlKey || event.metaKey) {
            return;
          }

          if (isComposing || event.nativeEvent.isComposing) {
            return;
          }

          if (!pendingEnter) {
            setPendingEnter(true);
            return;
          }

          event.preventDefault();
          setPendingEnter(false);

          const trimmed = value.trim();
          if (!trimmed || isSending || isUploadingAttachments) {
            return;
          }

          onSend(trimmed);
          setValue("");
          setIsUploadMenuOpen(false);
        }}
        placeholder={placeholder}
        ref={textareaRef}
        rows={1}
        value={value}
      />
      <div className="mt-4 flex items-center justify-between">
        <div className="relative">
          <input
            className="hidden"
            multiple
            onChange={(event) => {
              const files = Array.from(event.target.files ?? []);
              if (files.length > 0) {
                onUploadAttachments("file", files);
              }
              event.target.value = "";
              setIsUploadMenuOpen(false);
            }}
            ref={fileInputRef}
            type="file"
          />
          <input
            accept="image/*"
            className="hidden"
            multiple
            onChange={(event) => {
              const files = Array.from(event.target.files ?? []);
              if (files.length > 0) {
                onUploadAttachments("image", files);
              }
              event.target.value = "";
              setIsUploadMenuOpen(false);
            }}
            ref={imageInputRef}
            type="file"
          />
          {isUploadMenuOpen ? (
            <div className="absolute bottom-14 left-0 w-44 rounded-[24px] border border-line bg-white p-2 shadow-panel">
              <Button
                className="w-full justify-start rounded-[18px]"
                disabled={isUploadingAttachments}
                onClick={() => fileInputRef.current?.click()}
                type="button"
                variant="ghost"
              >
                <Paperclip className="mr-2 h-4 w-4" />
                上传文件
              </Button>
              <Button
                className="mt-1 w-full justify-start rounded-[18px]"
                disabled={isUploadingAttachments}
                onClick={() => imageInputRef.current?.click()}
                type="button"
                variant="ghost"
              >
                <ImagePlus className="mr-2 h-4 w-4" />
                上传图片
              </Button>
            </div>
          ) : null}
          <Button
            className="h-12 w-12 rounded-full p-0"
            disabled={isUploadingAttachments || isSending}
            onClick={() => setIsUploadMenuOpen((current) => !current)}
            size="icon"
            type="button"
            variant="secondary"
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>
        <Button className="h-12 w-12 rounded-full p-0" disabled={isSending || isUploadingAttachments || value.trim().length === 0} size="icon" type="submit">
          <ArrowUp className="h-4 w-4" />
        </Button>
      </div>
    </form>
  );
}
