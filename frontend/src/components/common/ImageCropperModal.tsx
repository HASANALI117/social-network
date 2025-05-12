import React, { useState, useRef } from 'react';
import ReactCrop, {
  centerCrop,
  makeAspectCrop,
  Crop,
  PixelCrop,
} from 'react-image-crop';
import 'react-image-crop/dist/ReactCrop.css';
import { Button } from '@/components/ui/button'; // Assuming you have a Button component
import { Dialog, DialogBody, DialogTitle, DialogActions } from '@/components/ui/dialog'; // Assuming ShadCN/UI Dialog

interface ImageCropperModalProps {
  isOpen: boolean;
  onClose: () => void;
  imageSrc: string | null;
  onCropComplete: (croppedImageBlob: Blob) => void;
  aspect?: number;
  circularCrop?: boolean;
}

// Utility function to get cropped image
async function getCroppedImg(
  imageElement: HTMLImageElement,
  crop: PixelCrop,
  fileName: string = 'cropped-image.png'
): Promise<Blob> {
  const canvas = document.createElement('canvas');
  const scaleX = imageElement.naturalWidth / imageElement.width;
  const scaleY = imageElement.naturalHeight / imageElement.height;

  canvas.width = Math.floor(crop.width * scaleX);
  canvas.height = Math.floor(crop.height * scaleY);

  const ctx = canvas.getContext('2d');
  if (!ctx) {
    return Promise.reject(new Error('Failed to get canvas context'));
  }

  const pixelRatio = window.devicePixelRatio || 1;
  canvas.width = Math.floor(crop.width * scaleX * pixelRatio);
  canvas.height = Math.floor(crop.height * scaleY * pixelRatio);
  ctx.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0);
  ctx.imageSmoothingQuality = 'high';

  ctx.drawImage(
    imageElement,
    crop.x * scaleX,
    crop.y * scaleY,
    crop.width * scaleX,
    crop.height * scaleY,
    0,
    0,
    crop.width * scaleX,
    crop.height * scaleY
  );

  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (!blob) {
          reject(new Error('Canvas is empty'));
          return;
        }
        // You can remove the fileName part if you don't need to create a File object here
        // const file = new File([blob], fileName, { type: blob.type });
        resolve(blob);
      },
      'image/png', // Or 'image/jpeg'
      0.95 // Quality for JPEG
    );
  });
}


const ImageCropperModal: React.FC<ImageCropperModalProps> = ({
  isOpen,
  onClose,
  imageSrc,
  onCropComplete,
  aspect = 1, // Default to square aspect ratio
  circularCrop = false, // Default to not circular
}) => {
  const [crop, setCrop] = useState<Crop | undefined>();
  const [completedCrop, setCompletedCrop] = useState<PixelCrop | undefined>();
  const imgRef = useRef<HTMLImageElement | null>(null);

  function onImageLoad(e: React.SyntheticEvent<HTMLImageElement>) {
    const { width, height } = e.currentTarget;
    const newCrop = centerCrop(
      makeAspectCrop(
        {
          unit: '%',
          width: 90, // Initial crop width
        },
        aspect,
        width,
        height
      ),
      width,
      height
    );
    setCrop(newCrop);
    // setCompletedCrop will be set by ReactCrop's onComplete handler
  }

  const handleCropImage = async () => {
    if (completedCrop && imgRef.current) {
      try {
        const croppedBlob = await getCroppedImg(imgRef.current, completedCrop);
        onCropComplete(croppedBlob);
        onClose();
      } catch (error) {
        console.error('Error cropping image:', error);
        // Handle error (e.g., show a notification to the user)
      }
    }
  };

  if (!isOpen || !imageSrc) {
    return null;
  }

  return (
    <Dialog open={isOpen} onClose={onClose}>
      <DialogBody className="sm:max-w-[500px]">
        <DialogTitle>Crop Image</DialogTitle>
        {imageSrc && (
          <div className="flex justify-center items-center my-4">
            <ReactCrop
              crop={crop}
              onChange={(_, percentCrop) => setCrop(percentCrop)}
              onComplete={(c) => setCompletedCrop(c)}
              aspect={aspect}
              circularCrop={circularCrop}
              minWidth={50} // Example: minimum crop width
              minHeight={50} // Example: minimum crop height
            >
              <img
                ref={imgRef}
                alt="Crop me"
                src={imageSrc}
                onLoad={onImageLoad}
                style={{ maxHeight: '70vh', objectFit: 'contain' }}
              />
            </ReactCrop>
          </div>
        )}
        <DialogActions>
          <Button outline onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleCropImage} disabled={!completedCrop?.width || !completedCrop?.height}>
            Crop Image
          </Button>
        </DialogActions>
      </DialogBody>
    </Dialog>
  );
};

export default ImageCropperModal;