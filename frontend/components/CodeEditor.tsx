"use client";

import Editor, { useMonaco } from "@monaco-editor/react";
import React, { useEffect } from "react";

import { editor as monacoEditor } from "monaco-editor";

interface CodeEditorProps {
  code?: string;
  output: string;
  onCodeChange: (code: string) => void;
}

const CodeEditor: React.FC<CodeEditorProps> = ({
  code = "",
  output,
  onCodeChange,
}) => {
  const monaco = useMonaco();

  useEffect(() => {
    if (!monaco) return;

    monaco.languages.register({ id: "custom" });

    const keywords = [
      "mkdisk",
      "rmdisk",
      "ls",
      "touch",
      "rm",
      "mv",
      "cp",
      "cat",
      "find",
      "grep",
      "chmod",
      "chown",
      "chgrp",
      "mkfile",
      "rep",
      "mkdir",
      "rmdir",
      "login",
      "logout",
      "mkuser",
      "rmuser",
      "mkgrp",
      "mount",
      "mounted",
      "fdisk",
      "loss",
      "unmount",
      "mkfs",
      "recovery",
      "journaling"
    ];

    const flags = [
      "path",
      "size",
      "fit",
      "ls_path",
      "unit",
      "type",
      "name",
      "id",
      "fs",
    ];

    monaco.languages.setMonarchTokensProvider("custom", {
      defaultToken: "",
      tokenPostfix: ".custom",

      tokenizer: {
        root: [
          // Líneas de error (comienzan con "error")
          [/^Error.*$/, "error"],

          // Comentarios (líneas que comienzan con #)
          [/#.*$/, "comment"],

          // Flags (palabras que comienzan con -)
          [/(-)([a-zA-Z0-9_]+)/, ["flag.prefix", "flag"]],

          // Keywords
          [new RegExp(`\\b(${keywords.join("|")})\\b`), "keyword"],

          // Cualquier otro texto
          [/[a-zA-Z0-9_]+/, "text"],
        ],
      },
    });

    // Definir un tema personalizado con colores más sutiles
    monaco.editor.defineTheme("custom-theme", {
      base: "vs-dark",
      inherit: true,
      rules: [
        { token: "error", foreground: "FF6B6B", fontStyle: "bold" }, // Rojo suave para errores
        { token: "comment", foreground: "6A9955", fontStyle: "italic" }, // Verde apagado para comentarios
        { token: "flag", foreground: "CE9178" }, // Naranja suave para flags
        { token: "flag.prefix", foreground: "CE9178" }, // Mismo color para el guión
        { token: "keyword", foreground: "569CD6" }, // Azul suave para keywords (sin negrita)
        { token: "text", foreground: "D4D4D4" }, // Gris claro para texto normal
      ],
      colors: {
        "editor.background": "#1E1E1E",
        "editor.lineHighlightBackground": "#2A2A2A",
      },
    });

    // Configuración para resaltar toda la línea de error
    monaco.editor.onDidCreateModel((model) => {
      const updateDecorations = () => {
        const decorations: unknown[] = [];
        const lines = model.getLinesContent();

        lines.forEach((line, index) => {
          if (line.trim().startsWith("Error")) {
            decorations.push({
              range: new monaco.Range(
                index + 1,
                1,
                index + 1,
                model.getLineMaxColumn(index + 1)
              ),
              options: {
                isWholeLine: true,
                className: "errorLine",
                glyphMarginClassName: "errorGlyphMargin",
              },
            });
          }
        });

        model.deltaDecorations([], decorations as []);
      };

      model.onDidChangeContent(updateDecorations);
      updateDecorations();
    });

    monaco.languages.registerCompletionItemProvider("custom", {
      provideCompletionItems: (model, position) => {
        const suggestions = [
          ...keywords.map((keyword) => ({
            label: keyword,
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertText: keyword,
            range: {
              startLineNumber: position.lineNumber,
              startColumn: position.column,
              endLineNumber: position.lineNumber,
              endColumn: position.column,
            },
          })),
          ...flags.map((flag) => ({
            label: `-${flag}`,
            kind: monaco.languages.CompletionItemKind.Property,
            insertText: `-${flag}`,
            range: {
              startLineNumber: position.lineNumber,
              startColumn: position.column,
              endLineNumber: position.lineNumber,
              endColumn: position.column,
            },
          })),
        ];
        return { suggestions };
      },
    });

    // Aplicar el tema
    monaco.editor.setTheme("custom-theme");

    // Añadir estilos CSS para el resaltado de línea de error
    const style = document.createElement("style");
    style.textContent = `
      .errorLine {
        background-color: rgba(255, 0, 0, 0.1);
      }
      .errorGlyphMargin {
        background-color: rgba(255, 0, 0, 0.3);
      }
    `;
    document.head.appendChild(style);

    return () => {
      document.head.removeChild(style);
    };
  }, [monaco]);

  // Función para resaltar líneas que empiecen con "Error" en un editor específico
  const highlightErrorLines = (
    editorInstance: monacoEditor.IStandaloneCodeEditor
  ) => {
    const model = editorInstance.getModel();
    if (!model) return;

    const updateDecorations = () => {
      const decorations: unknown[] = [];
      const lines = model.getLinesContent();

      lines.forEach((line: string, index: number) => {
        if (/^\s*(Error|error)/.test(line)) {
          // Busca "Error" o "error"
          decorations.push({
            range: new monaco!.Range(
              index + 1,
              1,
              index + 1,
              model.getLineMaxColumn(index + 1)
            ),
            options: {
              isWholeLine: true,
              className: "errorLine",
              glyphMarginClassName: "errorGlyphMargin",
            },
          });
        }
      });

      model.deltaDecorations([], decorations as []);
    };

    model.onDidChangeContent(updateDecorations);
    updateDecorations();
  };

  return (
    <div className="flex gap-4">
      <div className="rounded-xl overflow-hidden w-1/2">
        <Editor
          className="rounded-editor"
          height="62vh"
          defaultLanguage="custom"
          theme="custom-theme"
          options={{
            padding: {
              top: 25,
              bottom: 25,
            },
            glyphMargin: true, // Necesario para el resaltado de línea
          }}
          onChange={(value) => onCodeChange(value || "")}
          value={code}
        />
      </div>
      <div className="rounded-xl overflow-hidden w-1/2">
        <Editor
          className="rounded-editor"
          height="62vh"
          defaultLanguage="custom"
          theme="custom-theme"
          options={{
            readOnly: true,
            padding: {
              top: 25,
              bottom: 25,
            },
            glyphMargin: true,
          }}
          onMount={(editorInstance) => {
            highlightErrorLines(editorInstance);
          }}
          value={output}
        />
      </div>
    </div>
  );
};

export default CodeEditor;
