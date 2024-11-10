import React from 'react';
import {Image} from "@nextui-org/image";

const defaultImages: Record<string, string> = {
    'pi': '/pi.png',
    'k8s': '/k8s.png',
    'mock-truck': '/mock-truck.png',
    'flatbed-truck': '/flatbed-truck.png',
    'freezer-truck': '/freezer-truck.png',
    'box-truck': '/box-truck.png',
};

type CoverImageProps = {
  src: string;
  alt?: string;
  [key: string]: any; // For any additional props
};

const CoverImage = ({ src, alt = 'Cover image', ...props }: CoverImageProps) => {
    // Determine if src is a known name or a URL
    const imageUrl = defaultImages[src] || src;
  
    return (
      <Image
        removeWrapper                        
        alt={alt}
        style={{objectFit:"contain"}}
        className="z-0 absolute w-[30%] h-[30%] bottom-0 right-0"
        src={imageUrl}        
      />
    );
  };
  
  export default CoverImage;