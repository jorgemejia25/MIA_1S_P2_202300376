import React from "react";
import { DirectoryContent, FileContent, FileInfo } from "@/types/FileSystem";
import PageHeader from "../organisms/PageHeader";
import ErrorMessage from "../atoms/ErrorMessage";
import LoadingSpinner from "../atoms/LoadingSpinner";
import FileBreadcrumbs from "../organisms/FileBreadcrumbs";
import FileViewer from "../organisms/FileViewer";
import FilesContainer from "../organisms/FilesContainer";

interface PathSegment {
  name: string;
  path: string;
}

interface FilesPageTemplateProps {
  currentPath: string;
  directoryContent: DirectoryContent | null;
  loading: boolean;
  error: string | null;
  viewMode: "grid" | "list";
  selectedFile: FileInfo | null;
  fileContent: FileContent | null;
  loadingContent: boolean;
  pathSegments: PathSegment[];
  onPathChange: (newPath: string) => void;
  onFileClick: (file: FileInfo) => void;
  onCloseFileViewer: () => void;
  onRefresh: () => void;
  onViewModeChange: (mode: "grid" | "list") => void;
  onJournalingClick?: () => void;
}

const FilesPageTemplate: React.FC<FilesPageTemplateProps> = ({
  currentPath,
  directoryContent,
  loading,
  error,
  viewMode,
  selectedFile,
  fileContent,
  loadingContent,
  pathSegments,
  onPathChange,
  onFileClick,
  onCloseFileViewer,
  onRefresh,
  onViewModeChange,
  onJournalingClick,
}) => {
  const showFileViewer = selectedFile !== null;

  return (
    <div className="max-w-7xl mx-auto">
      <PageHeader
        title="Explorador de Archivos"
        description="Navega por el sistema de archivos y gestiona tus directorios y archivos."
      />

      {/* Barra de navegación con breadcrumbs y controles integrados */}
      <FileBreadcrumbs
        pathSegments={pathSegments}
        currentPath={currentPath}
        onPathChange={onPathChange}
        loading={loading}
        viewMode={viewMode}
        onRefresh={onRefresh}
        onViewModeChange={onViewModeChange}
        onJournalingClick={onJournalingClick}
      />

      <ErrorMessage message={error} />

      {loading && <LoadingSpinner />}

      {/* Contenedor principal */}
      {!showFileViewer ? (
        <div className="bg-neutral-800/60 backdrop-blur-sm rounded-xl overflow-hidden p-5 ring-1 ring-neutral-700">
          <FilesContainer
            directoryContent={directoryContent}
            viewMode={viewMode}
            loading={loading}
            onFileClick={onFileClick}
            onRefresh={onRefresh}
          />
        </div>
      ) : (
        selectedFile && (
          <FileViewer
            selectedFile={selectedFile}
            fileContent={fileContent}
            loadingContent={loadingContent}
            onClose={onCloseFileViewer}
          />
        )
      )}
    </div>
  );
};

export default FilesPageTemplate;
