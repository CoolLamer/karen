import SwiftUI

struct VIPContactsStepView: View {
    @ObservedObject var viewModel: OnboardingViewModel
    @State private var newVipName = ""

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                VStack(spacing: 8) {
                    Text("Koho má Karen vždy přepojit?")
                        .font(.title2)
                        .fontWeight(.bold)

                    Text("Některé hovory jsou důležité a nechceš, aby je Karen vyřizovala.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }
                .padding(.top, 20)

                // Explanation
                VStack(alignment: .leading, spacing: 8) {
                    Text("Když se volající představí jedním z těchto jmen, Karen ho okamžitě přepojí na tebe.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)

                    Text("Například: \"Tady máma\" -> Karen řekne \"Přepojuji\" a zavolá ti.")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .italic()
                }
                .padding()
                .background(Color(.systemBackground))
                .clipShape(RoundedRectangle(cornerRadius: 12))

                // VIP names input
                VStack(alignment: .leading, spacing: 12) {
                    Text("VIP jména")
                        .font(.headline)

                    HStack {
                        TextField("Máma, Táta, Jana...", text: $newVipName)
                            .textFieldStyle(.roundedBorder)
                            .onSubmit {
                                addVipName()
                            }

                        Button {
                            addVipName()
                        } label: {
                            Image(systemName: "plus.circle.fill")
                                .font(.title2)
                        }
                        .disabled(newVipName.isEmpty)
                    }

                    // Tags display
                    if !viewModel.vipNames.isEmpty {
                        FlowLayout(spacing: 8) {
                            ForEach(Array(viewModel.vipNames.enumerated()), id: \.offset) { index, name in
                                HStack(spacing: 4) {
                                    Text(name)
                                        .font(.subheadline)
                                    Button {
                                        viewModel.removeVipName(at: index)
                                    } label: {
                                        Image(systemName: "xmark.circle.fill")
                                            .font(.caption)
                                    }
                                }
                                .padding(.horizontal, 12)
                                .padding(.vertical, 6)
                                .background(Color.accentColor.opacity(0.1))
                                .foregroundStyle(Color.accentColor)
                                .clipShape(Capsule())
                            }
                        }
                    }
                }
                .padding()
                .background(Color(.systemBackground))
                .clipShape(RoundedRectangle(cornerRadius: 12))

                Spacer(minLength: 20)

                HStack(spacing: 12) {
                    Button {
                        viewModel.goToNext()
                    } label: {
                        Text("Nastavím později")
                            .frame(maxWidth: .infinity)
                            .padding()
                            .foregroundStyle(.secondary)
                    }

                    Button {
                        viewModel.goToNext()
                    } label: {
                        HStack {
                            Text("Pokračovat")
                            Image(systemName: "arrow.right")
                        }
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.accentColor)
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                    }
                }
            }
            .padding()
        }
    }

    private func addVipName() {
        let trimmed = newVipName.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return }
        viewModel.addVipName(trimmed)
        newVipName = ""
    }
}

// MARK: - Flow Layout

struct FlowLayout: Layout {
    var spacing: CGFloat = 8

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let sizes = subviews.map { $0.sizeThatFits(.unspecified) }
        return layout(sizes: sizes, containerWidth: proposal.width ?? .infinity).size
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let sizes = subviews.map { $0.sizeThatFits(.unspecified) }
        let offsets = layout(sizes: sizes, containerWidth: bounds.width).offsets

        for (offset, subview) in zip(offsets, subviews) {
            subview.place(at: CGPoint(x: bounds.minX + offset.x, y: bounds.minY + offset.y), proposal: .unspecified)
        }
    }

    private func layout(sizes: [CGSize], containerWidth: CGFloat) -> (offsets: [CGPoint], size: CGSize) {
        var offsets: [CGPoint] = []
        var currentX: CGFloat = 0
        var currentY: CGFloat = 0
        var lineHeight: CGFloat = 0
        var maxWidth: CGFloat = 0

        for size in sizes {
            if currentX + size.width > containerWidth && currentX > 0 {
                currentX = 0
                currentY += lineHeight + spacing
                lineHeight = 0
            }

            offsets.append(CGPoint(x: currentX, y: currentY))
            currentX += size.width + spacing
            lineHeight = max(lineHeight, size.height)
            maxWidth = max(maxWidth, currentX)
        }

        return (offsets, CGSize(width: maxWidth, height: currentY + lineHeight))
    }
}

#Preview {
    VIPContactsStepView(viewModel: OnboardingViewModel())
}
