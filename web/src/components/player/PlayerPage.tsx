import React, {useContext, useState, useEffect} from "react";
import Player from "./Player";
import {useNavigate} from "react-router-dom";
import {CinemaModeContext} from "../../providers/CinemaModeProvider";
import ModalTextInput from "../shared/ModalTextInput";
import ChatBox from "../chat/ChatBox";

const PlayerPage = () => {
  const navigate = useNavigate();
  const {cinemaMode, toggleCinemaMode} = useContext(CinemaModeContext);
  const [streamKeys, setStreamKeys] = useState<string[]>([window.location.pathname.substring(1)]);
  const [isModalOpen, setIsModelOpen] = useState<boolean>(false);
  const [isFullscreen, setIsFullscreen] = useState<boolean>(false);
  const [isStreamLive, setIsStreamLive] = useState<boolean>(false);

  const addStream = (streamKey: string) => {
    if (streamKeys.some((key: string) => key.toLowerCase() === streamKey.toLowerCase())) {
      return;
    }
    setStreamKeys((prev) => [...prev, streamKey]);
    setIsModelOpen((prev) => !prev);
  };

  // Listen for video playing events to detect if stream is live
  useEffect(() => {
    const handleVideoPlaying = () => setIsStreamLive(true);
    const handleVideoEnded = () => setIsStreamLive(false);

    // Listen for the custom playing event from video elements
    const videoElement = document.querySelector('video');
    if (videoElement) {
      videoElement.addEventListener('playing', handleVideoPlaying);
      videoElement.addEventListener('ended', handleVideoEnded);
      videoElement.addEventListener('emptied', handleVideoEnded);

      return () => {
        videoElement.removeEventListener('playing', handleVideoPlaying);
        videoElement.removeEventListener('ended', handleVideoEnded);
        videoElement.removeEventListener('emptied', handleVideoEnded);
      };
    }
  }, [streamKeys]);

  // Detect fullscreen changes
  useEffect(() => {
    const handleFullscreenChange = () => {
      const isNowFullscreen = !!(
        document.fullscreenElement ||
        (document as any).webkitFullscreenElement ||
        (document as any).mozFullScreenElement ||
        (document as any).msFullscreenElement
      );
      setIsFullscreen(isNowFullscreen);
    };

    document.addEventListener('fullscreenchange', handleFullscreenChange);
    document.addEventListener('webkitfullscreenchange', handleFullscreenChange);
    document.addEventListener('mozfullscreenchange', handleFullscreenChange);
    document.addEventListener('msfullscreenchange', handleFullscreenChange);

    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
      document.removeEventListener('webkitfullscreenchange', handleFullscreenChange);
      document.removeEventListener('mozfullscreenchange', handleFullscreenChange);
      document.removeEventListener('msfullscreenchange', handleFullscreenChange);
    };
  }, []);

  return (
    <div>
      {isModalOpen && (
        <ModalTextInput<string>
          title="Add stream"
          message={"Insert stream key to add to multi stream"}
          isOpen={isModalOpen}
          canCloseOnBackgroundClick={false}
          onClose={() => setIsModelOpen(false)}
          onAccept={(result: string) => addStream(result)}
        />
      )}

      <div className={`flex flex-col ${streamKeys.length === 1 && !cinemaMode && isStreamLive ? 'md:flex-row' : ''} w-full ${!cinemaMode && "mx-auto px-2 py-2 container gap-2"}`}>
        <div className={`w-full ${streamKeys.length === 1 && !cinemaMode && isStreamLive ? 'md:flex-1' : ''}`}>
          <div className={`grid ${streamKeys.length !== 1 ? "grid-cols-1 md:grid-cols-2" : ""}  w-full gap-2`}>
            {streamKeys.map((streamKey) =>
              <Player
                key={`${streamKey}_player`}
                streamKey={streamKey}
                cinemaMode={cinemaMode}
                onCloseStream={
                  streamKeys.length === 1
                    ? () => navigate('/')
                    : () => setStreamKeys((prev) => prev.filter((key) => key !== streamKey))
                }
              />
            )}
          </div>
        </div>

        {/* Chat - Only show when stream is live */}
        {streamKeys.length === 1 && isStreamLive && (
          <>
            {/* Desktop side panel */}
            {!cinemaMode && !isFullscreen && (
              <div className="hidden md:block w-80 shrink-0">
                <ChatBox streamKey={streamKeys[0]} />
              </div>
            )}

            {/* Fullscreen overlay (desktop and mobile) */}
            {isFullscreen && (
              <ChatBox
                streamKey={streamKeys[0]}
                isFullscreen={true}
              />
            )}

            {/* Mobile slide-up (non-fullscreen) */}
            {!cinemaMode && !isFullscreen && (
              <div className="md:hidden mt-4">
                <ChatBox streamKey={streamKeys[0]} className="h-96" />
              </div>
            )}
          </>
        )}

        {/*Implement footer menu*/}
        <div className="flex flex-row p-2 gap-2">
          <button
            className="bg-blue-900 hover:bg-blue-800 px-4 py-2 rounded-lg mt-6"
            onClick={toggleCinemaMode}
          >
            {cinemaMode ? "Disable cinema mode" : "Enable cinema mode"}
          </button>

          {/*Show modal to add stream keys with*/}
          <button
            className="bg-blue-900 hover:bg-blue-800 px-4 py-2 rounded-lg mt-6"
            onClick={() => setIsModelOpen((prev) => !prev)}>
            Add Stream
          </button>
        </div>
      </div>
    </div>
  )
};

export default PlayerPage;