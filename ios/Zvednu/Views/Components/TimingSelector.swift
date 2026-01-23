import SwiftUI

/// A reusable timing selector component for no-answer redirect configuration.
/// Displays a horizontal scrollable list of time options (5s, 10s, etc.)
struct TimingSelector: View {
    @Binding var selectedTime: Int
    var showLabel: Bool = true

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            if showLabel {
                Text("Po kolika sekundách přesměrovat?")
                    .font(.caption)
                    .fontWeight(.medium)
            }

            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 8) {
                    ForEach(RedirectCodes.noAnswerTimeOptions, id: \.self) { time in
                        Button {
                            selectedTime = time
                        } label: {
                            Text("\(time)s")
                                .font(.caption)
                                .padding(.horizontal, 12)
                                .padding(.vertical, 8)
                                .background(
                                    selectedTime == time
                                        ? Color.accentColor
                                        : Color(.systemGray5)
                                )
                                .foregroundStyle(
                                    selectedTime == time
                                        ? .white
                                        : .primary
                                )
                                .clipShape(Capsule())
                        }
                    }
                }
            }
        }
    }
}

#Preview {
    TimingSelector(selectedTime: .constant(10))
        .padding()
}
