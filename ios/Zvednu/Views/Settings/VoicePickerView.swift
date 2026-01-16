import SwiftUI
import AVFoundation

struct VoicePickerView: View {
    @ObservedObject var viewModel: SettingsViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var selectedVoiceId: String = ""
    @State private var playingVoiceId: String?
    @State private var loadingVoiceId: String?
    @State private var audioPlayer: AVAudioPlayer?
    @State private var audioDelegate: AudioPlayerDelegateHandler?

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoadingVoices && viewModel.voices.isEmpty {
                    ProgressView("Načítám hlasy...")
                } else {
                    voiceList
                }
            }
            .navigationTitle("Vyberte hlas")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button("Zrušit") {
                        stopAudio()
                        dismiss()
                    }
                }
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Potvrdit") {
                        Task {
                            stopAudio()
                            await viewModel.selectVoice(selectedVoiceId)
                        }
                    }
                    .disabled(selectedVoiceId.isEmpty || viewModel.isSaving)
                }
            }
            .task {
                await viewModel.loadVoices()
                // Set initial selection
                if let currentVoiceId = viewModel.tenant?.voiceId {
                    selectedVoiceId = currentVoiceId
                } else if let firstVoice = viewModel.voices.first {
                    selectedVoiceId = firstVoice.id
                }
            }
            .onDisappear {
                stopAudio()
            }
        }
    }

    private var voiceList: some View {
        List {
            ForEach(viewModel.voices) { voice in
                VoiceRowView(
                    voice: voice,
                    isSelected: selectedVoiceId == voice.id,
                    isPlaying: playingVoiceId == voice.id,
                    isLoading: loadingVoiceId == voice.id,
                    onSelect: {
                        selectedVoiceId = voice.id
                    },
                    onPreview: {
                        Task {
                            await previewVoice(voice.id)
                        }
                    }
                )
            }

            Section {
                Text("Vybraný hlas se použije při příštím hovoru.")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
    }

    private func previewVoice(_ voiceId: String) async {
        // Stop any currently playing audio
        stopAudio()

        // If clicking the same voice that was playing, just stop
        if playingVoiceId == voiceId {
            playingVoiceId = nil
            return
        }

        loadingVoiceId = voiceId

        if let audioData = await viewModel.previewVoice(voiceId) {
            do {
                // Configure audio session
                try AVAudioSession.sharedInstance().setCategory(.playback, mode: .default)
                try AVAudioSession.sharedInstance().setActive(true)

                audioPlayer = try AVAudioPlayer(data: audioData)

                // Create and store the delegate to prevent deallocation
                let delegate = AudioPlayerDelegateHandler {
                    Task { @MainActor in
                        self.playingVoiceId = nil
                    }
                }
                audioDelegate = delegate
                audioPlayer?.delegate = delegate
                audioPlayer?.play()
                playingVoiceId = voiceId
            } catch {
                viewModel.error = "Nepodařilo se přehrát ukázku"
            }
        }

        loadingVoiceId = nil
    }

    private func stopAudio() {
        audioPlayer?.stop()
        audioPlayer = nil
        playingVoiceId = nil
    }
}

// MARK: - Voice Row View

struct VoiceRowView: View {
    let voice: Voice
    let isSelected: Bool
    let isPlaying: Bool
    let isLoading: Bool
    let onSelect: () -> Void
    let onPreview: () -> Void

    var body: some View {
        Button(action: onSelect) {
            HStack(spacing: 12) {
                // Selection indicator
                Image(systemName: isSelected ? "checkmark.circle.fill" : "circle")
                    .foregroundStyle(isSelected ? .teal : .secondary)
                    .font(.title3)

                // Gender icon
                Image(systemName: voice.gender == "female" ? "person.fill" : "person.fill")
                    .foregroundStyle(voice.gender == "female" ? .pink : .blue)
                    .frame(width: 24)

                // Voice info
                VStack(alignment: .leading, spacing: 2) {
                    Text(voice.name)
                        .font(.body)
                        .fontWeight(.medium)
                        .foregroundStyle(.primary)
                    Text(voice.description)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                Spacer()

                // Preview button
                Button(action: onPreview) {
                    if isLoading {
                        ProgressView()
                            .frame(width: 32, height: 32)
                    } else {
                        Image(systemName: isPlaying ? "stop.fill" : "play.fill")
                            .font(.body)
                            .foregroundStyle(isPlaying ? .teal : .secondary)
                            .frame(width: 32, height: 32)
                    }
                }
                .buttonStyle(.plain)
                .disabled(isLoading)
            }
            .padding(.vertical, 4)
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Audio Player Delegate

final class AudioPlayerDelegateHandler: NSObject, AVAudioPlayerDelegate, @unchecked Sendable {
    private let onFinish: @Sendable () -> Void

    init(onFinish: @escaping @Sendable () -> Void) {
        self.onFinish = onFinish
    }

    func audioPlayerDidFinishPlaying(_ player: AVAudioPlayer, successfully flag: Bool) {
        onFinish()
    }
}

#Preview {
    VoicePickerView(viewModel: SettingsViewModel())
}
