import { defineToolcraft } from "@/toolcraft/runtime";

export const appSchema = defineToolcraft({
  canvas: {
    draggable: true,
    enabled: true,
    size: { height: 1080, unit: "px", width: 1920 },
    sizing: { mode: "editable-output" },
    upload: true,
  },
  media: {
    defaultAssets: [
      {
        assetKind: "image",
        dataUrl: "/img.jpg",
        fileName: "img.jpg",
        id: "duck-source",
        mimeType: "image/jpeg",
        sourceTarget: "source.image",
      },
    ],
  },
  panels: {
    controls: {
      sections: [
        {
          controls: {
            variant: {
              defaultValue: "hero",
              label: "Variant",
              options: [
                { label: "Hero", value: "hero" },
                { label: "Thinking", value: "thinking" },
                { label: "Chat", value: "chat" },
                { label: "Cards", value: "cards" },
              ],
              orderRole: "mode",
              performanceReason: "Switching layout recomposes a small DOM card without changing source decode cost.",
              performanceRole: "responsiveness",
              target: "card.variant",
              type: "segmented",
            },
          },
          title: "Layout",
        },
        {
          controls: {
            image: {
              assetKind: "image",
              defaultValue: null,
              label: "Photo",
              orderRole: "input",
              performanceReason: "Replacing the single source image must keep the editor responsive.",
              performanceRole: "responsiveness",
              target: "source.image",
              type: "fileDrop",
            },
            imageScale: {
              defaultValue: 104,
              label: "Image scale",
              max: 140,
              min: 80,
              orderRole: "spatial",
              performanceReason: "Image scale changes live crop geometry while preserving the decoded source.",
              performanceRole: "workload",
              step: 1,
              target: "source.scale",
              type: "slider",
              unit: "%",
            },
          },
          title: "Source",
        },
        {
          controls: {
            headline: {
              commitMode: "content",
              defaultValue: "Утёнок думает за вас",
              label: "Headline",
              orderRole: "input",
              performanceReason: "Headline updates native preview text and lightweight export text layout.",
              performanceRole: "responsiveness",
              target: "copy.headline",
              type: "text",
            },
            caption: {
              commitMode: "content",
              defaultValue: "Ваш дружелюбный AI-агент в Telegram",
              label: "Caption",
              orderRole: "input",
              performanceReason: "Caption updates native preview text and lightweight export text layout.",
              performanceRole: "responsiveness",
              target: "copy.caption",
              type: "text",
            },
            handle: {
              commitMode: "content",
              defaultValue: "@duck_agent",
              label: "Telegram",
              orderRole: "input",
              performanceReason: "Handle updates one short native text label.",
              performanceRole: "responsiveness",
              target: "copy.handle",
              type: "text",
            },
          },
          title: "Message",
        },
        {
          controls: {
            accent: {
              defaultValue: "#229ED9",
              label: "Accent",
              orderRole: "color",
              performanceReason: "Accent changes CSS color tokens without increasing render workload.",
              performanceRole: "responsiveness",
              target: "appearance.accent",
              type: "color",
            },
          },
          title: "Brand",
        },
        {
          controls: {
            includeBackground: {
              defaultValue: true,
              description: "Controls live preview and PNG/JPG background visibility.",
              label: "Include",
              orderRole: "primary",
              performanceReason: "The switch only toggles the product background layer.",
              performanceRole: "responsiveness",
              target: "export.includeBackground",
              type: "switch",
            },
            background: {
              defaultValue: "#F4F1E8",
              label: false,
              orderRole: "color",
              performanceReason: "Background color changes one CSS token and the export fill.",
              performanceRole: "responsiveness",
              target: "appearance.background",
              type: "color",
            },
          },
          layoutGroups: [
            {
              columns: 2,
              controls: ["includeBackground", "background"],
              layout: "inline",
            },
          ],
          title: "Background",
        },
        {
          controls: {
            imageFormat: {
              defaultValue: "png",
              label: "Format",
              options: [
                { label: "PNG", value: "png" },
                { label: "JPG", value: "jpg" },
              ],
              orderRole: "advanced",
              performanceReason: "Format selection changes only encoding at export time.",
              performanceRole: "responsiveness",
              target: "export.image.format",
              type: "select",
            },
            imageResolution: {
              defaultValue: "4k",
              label: "Resolution",
              options: [
                { label: "2K", value: "2k" },
                { label: "4K", value: "4k" },
                { label: "8K", value: "8k" },
              ],
              orderRole: "advanced",
              performanceReason: "Selected long-edge resolution controls export canvas allocation and composite cost.",
              performanceRole: "workload",
              target: "export.image.resolution",
              type: "select",
            },
          },
          layoutGroups: [
            {
              columns: 2,
              controls: ["imageFormat", "imageResolution"],
              layout: "inline",
            },
          ],
          title: "Image Export",
        },
        {
          actionGroup: "primary",
          controls: {
            outputActions: {
              actions: [
                {
                  icon: "upload-simple",
                  label: "Export image",
                  value: "export.png",
                },
              ],
              target: "actions.output",
              type: "panelActions",
            },
          },
        },
      ],
      title: "Duck Agent Cards",
    },
  },
  persistence: { storage: "none" },
  settingsTransfer: {
    appId: "duck-agent-cards",
    enabled: "auto",
    fileName: "duck-agent-cards-settings",
  },
  toolbar: {
    history: true,
    radar: true,
    theme: true,
    zoom: true,
  },
});
