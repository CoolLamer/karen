import SwiftUI

struct TranscriptBubbleView: View {
    let utterance: Utterance

    var body: some View {
        HStack {
            if utterance.isAgent {
                Spacer(minLength: 40)
            }

            VStack(alignment: utterance.isAgent ? .trailing : .leading, spacing: 4) {
                // Speaker label
                HStack(spacing: 4) {
                    Image(systemName: utterance.isAgent ? "phone.badge.checkmark" : "person.fill")
                        .font(.caption2)
                    Text(utterance.speakerDisplayName)
                        .font(.caption)
                        .fontWeight(.medium)
                }
                .foregroundStyle(utterance.isAgent ? Color.accentColor : Color.secondary)

                // Message bubble
                Text(utterance.text)
                    .font(.subheadline)
                    .padding(12)
                    .background(utterance.isAgent ? Color.accentColor.opacity(0.1) : Color(.systemGray6))
                    .clipShape(
                        RoundedRectangle(cornerRadius: 16)
                    )
            }

            if !utterance.isAgent {
                Spacer(minLength: 40)
            }
        }
    }
}

#Preview {
    VStack(spacing: 12) {
        TranscriptBubbleView(utterance: Utterance(
            speaker: "agent",
            text: "Dobry den, tady asistentka Karen. Lukas ted nemuze prijmout hovor, ale muzu vam pro nej zanechat vzkaz - co od nej potrebujete?",
            sequence: 1,
            startedAt: nil,
            endedAt: nil,
            sttConfidence: nil,
            interrupted: false
        ))

        TranscriptBubbleView(utterance: Utterance(
            speaker: "caller",
            text: "Dobry den, volam ohledne nabidky pojisteni.",
            sequence: 2,
            startedAt: nil,
            endedAt: nil,
            sttConfidence: nil,
            interrupted: false
        ))

        TranscriptBubbleView(utterance: Utterance(
            speaker: "agent",
            text: "Dekuji za zajem, ale Lukas si momentalne nepeje nove nabidky pojisteni. Mejte se hezky, nashledanou.",
            sequence: 3,
            startedAt: nil,
            endedAt: nil,
            sttConfidence: nil,
            interrupted: false
        ))
    }
    .padding()
}
