import React from "react";

interface PageHeaderProps {
  title: string;
  description: string;
}

const PageHeader: React.FC<PageHeaderProps> = ({ title, description }) => {
  return (
    <header className="my-8">
      <h1 className="text-5xl font-bold mb-4 bg-clip-text text-white">
        {title}
      </h1>
      <p className="text-gray-400 text-sm">{description}</p>
    </header>
  );
};

export default PageHeader;
