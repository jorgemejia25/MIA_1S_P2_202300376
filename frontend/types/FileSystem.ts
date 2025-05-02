export interface FileInfo {
  name: string; // Nombre del archivo o carpeta
  type: "file" | "directory"; // Tipo de entrada
  size: number; // Tamaño en bytes
  permissions: string; // Permisos en formato octal (ej. "777")
  owner: number; // ID del propietario
  group: number; // ID del grupo
  modTime: string; // Fecha de modificación
  inodeId: number; // ID del inodo
}

export interface DirectoryContent {
  path: string; // Ruta del directorio listado
  files: FileInfo[]; // Lista de archivos en el directorio
  success: boolean; // Indicador de éxito de la operación
  errorMsg?: string; // Mensaje de error si ocurre alguno
}

export interface DirectoryLsResponse {
  success: boolean;
  message?: string;
  content?: DirectoryContent;
}

export interface FileContent {
  success: boolean;
  message?: string;
  name: string;
  content: string;
}
