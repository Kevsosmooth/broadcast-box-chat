import React, {createRef, useState} from 'react'
import {useNavigate} from 'react-router-dom'
import AvailableStreams from "./AvailableStreams";

const Frontpage = () => {
	const [streamType, setStreamType] = useState<'Watch' | 'Share'>('Watch');
	const streamKey = createRef<HTMLInputElement>()
	const navigate = useNavigate()

	const onStreamClick = () => {
		if(!streamKey.current || streamKey.current?.value === ''){
			return;
		}
		
		if(streamType === "Share"){
			navigate(`/publish/${streamKey.current.value}`)
		}

		if(streamType === "Watch"){
			navigate(`/${streamKey.current.value}`)
		}
	}

	return (
		<div className='space-y-4 mx-auto max-w-2xl pt-12 md:pt-20 px-4 md:px-0'>
			<div className='rounded-md bg-gray-800 shadow-md p-4 md:p-8'>
				<h2 className="font-light leading-tight text-2xl md:text-4xl mt-0 mb-2">Welcome to Broadcast Box</h2>
				<p className="text-sm md:text-base">Broadcast Box is a tool that allows you to efficiently stream high-quality video in real time, using the latest in video codecs and WebRTC technology.</p>

				<div className="flex flex-col md:flex-row rounded-md shadow-xs justify-center mt-6 gap-2 md:gap-0" role="group">

					<button
						type="button"
						onClick={() => setStreamType('Watch')}
						className={`flex items-center justify-center px-4 py-3 md:py-2 text-sm md:text-base font-medium border border-gray-200 rounded-lg md:rounded-s-lg md:rounded-e-none hover:text-blue-700 dark:border-gray-700 dark:text-white dark:hover:text-white dark:hover:bg-blue-700 dark:focus:ring-blue-500 dark:focus:text-white transition-colors ${streamType === "Watch" ? "bg-blue-700" : ""}`}>
						<svg className="w-5 h-5 md:w-6 md:h-6 mr-2 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" viewBox="0 0 24 24">
							<path stroke="currentColor" strokeLinecap="round" strokeWidth="2" d="M4.5 17H4a1 1 0 0 1-1-1 3 3 0 0 1 3-3h1m0-3.05A2.5 2.5 0 1 1 9 5.5M19.5 17h.5a1 1 0 0 0 1-1 3 3 0 0 0-3-3h-1m0-3.05a2.5 2.5 0 1 0-2-4.45m.5 13.5h-7a1 1 0 0 1-1-1 3 3 0 0 1 3-3h3a3 3 0 0 1 3 3 1 1 0 0 1-1 1Zm-1-9.5a2.5 2.5 0 1 1-5 0 2.5 2.5 0 0 1 5 0Z"/>
						</svg>
						I want to watch
					</button>
					<button
						type="button"
						onClick={() => setStreamType('Share')}
						className={`flex items-center justify-center px-4 py-3 md:py-2 text-sm md:text-base font-medium border border-gray-200 rounded-lg md:rounded-e-lg md:rounded-s-none hover:text-blue-700 dark:border-gray-700 dark:text-white dark:hover:text-white dark:hover:bg-blue-700 dark:focus:ring-blue-500 dark:focus:text-white transition-colors ${streamType === "Share" ? "bg-blue-700" : ""}`}>
						<svg className="w-5 h-5 md:w-6 md:h-6 mr-2 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" viewBox="0 0 24 24">
							<path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M14 6H4a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h10a1 1 0 0 0 1-1V7a1 1 0 0 0-1-1Zm7 11-6-2V9l6-2v10Z"/>
						</svg>
						I want to stream
					</button>

				</div>

					<div className='flex flex-col my-4 md:my-6 justify-center'>
						<label className='block text-sm md:text-base font-bold mb-2' htmlFor='streamKey'>
							Stream Key
						</label>

						<input
							className='mb-3 md:mb-2 appearance-none border w-full py-3 md:py-2 px-4 md:px-3 text-base md:text-sm leading-tight focus:outline-none focus:shadow-outline focus:ring-2 focus:ring-blue-500 bg-gray-700 border-gray-700 text-white rounded-lg shadow-md placeholder-gray-400'
							id='streamKey'
							placeholder={`Insert the key of the stream you want to ${streamType === "Share" ? 'share' : 'join'}`}
							type='text'
							onKeyUp={(e => {
								if(e.key === "Enter"){
									onStreamClick()
								}
							})}
							ref={streamKey}
							autoFocus/>

						<button
							className={`py-3 md:py-2 px-4 text-base md:text-sm ${streamKey.current?.value.length === 0 ? "bg-gray-700" : "bg-blue-600"} text-white font-semibold rounded-lg shadow-md ${streamKey.current?.value.length === 0 ? "hover:bg-gray-600" : "hover:bg-blue-700" } focus:outline-none focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75 transition-colors disabled:opacity-50 disabled:cursor-not-allowed`}
							disabled={streamKey.current?.value.length === 0}
							type='button'
							onClick={onStreamClick}>
							{streamType === "Share" ? "Start stream" : "Join stream"}
						</button>
					</div>

				<AvailableStreams/>
			</div>
		</div>
	)
}

export default Frontpage