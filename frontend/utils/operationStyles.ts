export type OperationStyleConfig = {
  bg: string;
  text: string;
  border: string;
  iconClass: string;
  iconBg: string;
  hoverBg: string;
  badgeBg: string;
};

export const getOperationConfig = (operation: string): OperationStyleConfig => {
  const configs: Record<string, OperationStyleConfig> = {
    mkdir: {
      bg: "bg-blue-900/10",
      text: "text-blue-300",
      border: "border-blue-700/30",
      iconClass: "text-blue-400",
      iconBg: "bg-blue-600/10",
      hoverBg: "hover:bg-blue-900/20",
      badgeBg: "bg-blue-500/10",
    },

    mkfile: {
      bg: "bg-emerald-900/10",
      text: "text-emerald-300",
      border: "border-emerald-700/30",
      iconClass: "text-emerald-400",
      iconBg: "bg-emerald-600/10",
      hoverBg: "hover:bg-emerald-900/20",
      badgeBg: "bg-emerald-500/10",
    },

    chmod: {
      bg: "bg-amber-900/10",
      text: "text-amber-300",
      border: "border-amber-700/30",
      iconClass: "text-amber-400",
      iconBg: "bg-amber-600/10",
      hoverBg: "hover:bg-amber-900/20",
      badgeBg: "bg-amber-500/10",
    },

    append: {
      bg: "bg-purple-900/10",
      text: "text-purple-300",
      border: "border-purple-700/30",
      iconClass: "text-purple-400",
      iconBg: "bg-purple-600/10",
      hoverBg: "hover:bg-purple-900/20",
      badgeBg: "bg-purple-500/10",
    },

    rm: {
      bg: "bg-red-900/10",
      text: "text-red-300",
      border: "border-red-700/30",
      iconClass: "text-red-400",
      iconBg: "bg-red-600/10",
      hoverBg: "hover:bg-red-900/20",
      badgeBg: "bg-red-500/10",
    },

    rename: {
      bg: "bg-cyan-900/10",
      text: "text-cyan-300",
      border: "border-cyan-700/30",
      iconClass: "text-cyan-400",
      iconBg: "bg-cyan-600/10",
      hoverBg: "hover:bg-cyan-900/20",
      badgeBg: "bg-cyan-500/10",
    },

    default: {
      bg: "bg-gray-800/30",
      text: "text-gray-300",
      border: "border-gray-600/30",
      iconClass: "text-gray-400",
      iconBg: "bg-gray-700/20",
      hoverBg: "hover:bg-gray-700/30",
      badgeBg: "bg-gray-700/20",
    },
  };

  const operationType = operation in configs ? operation : "default";
  return configs[operationType];
};
